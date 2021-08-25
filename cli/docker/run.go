package docker

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type RunOpts struct {
	Name       string
	ImageTags  string
	WorkingDir string
	Cmd        []string

	Volumes map[string]string

	ShowOutput bool
	Debug      bool
}

func waitForContainer(cli *client.Client, containerID string, showOutput bool) error {
	var err error

	var wg sync.WaitGroup
	c := make(common.ReadyChan, 1)

	var outputBuf *os.File
	var out io.Writer

	ctx := context.Background()
	logsReader, err := cli.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		Follow:     true,
	})
	if err != nil {
		return fmt.Errorf("Failed to check container logs: %s", err)
	}

	if showOutput {
		out = os.Stdout
	} else {
		if outputBuf, err = ioutil.TempFile("", "out"); err != nil {
			out = ioutil.Discard
			log.Warnf("Failed to create tmp file to store docker run output: %s", err)
		} else {
			out = outputBuf
			defer outputBuf.Close()
			defer os.Remove(outputBuf.Name())
		}

		wg.Add(1)
		go common.StartCommandSpinner(c, &wg, "")
	}

	wg.Add(1)
	go func(buildErr *error) {
		defer wg.Done()
		defer common.SendReady(c)

		if _, err := io.Copy(out, logsReader); err != nil {
			*buildErr = err
			return
		}
	}(&err)

	wg.Wait()

	statusCh, errCh := cli.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			if outputBuf != nil {
				if err := common.PrintFromStart(outputBuf); err != nil {
					log.Warnf("Failed to show docker run output: %s", err)
				}
			}
			return fmt.Errorf("Failed to wait for container to stop: %s", err)
		}
	case statusBody := <-statusCh:
		if statusBody.StatusCode != 0 {
			if outputBuf != nil {
				if err := common.PrintFromStart(outputBuf); err != nil {
					log.Warnf("Failed to show docker build output: %s", err)
				}
			}
			return fmt.Errorf("exited with code %d", statusBody.StatusCode)
		}
	}

	if err != nil {
		if outputBuf != nil {
			if err := common.PrintFromStart(outputBuf); err != nil {
				log.Warnf("Failed to show docker run output: %s", err)
			}
		}

		return err
	}

	return nil
}

func RunContainer(opts RunOpts) error {
	cli, err := client.NewClientWithOpts()
	if err != nil {
		return err
	}

	// init volumes
	var binds []string
	for hostPath, containerPath := range opts.Volumes {
		binds = append(binds, fmt.Sprintf("%s:%s", hostPath, containerPath))
	}

	ctx := context.Background()
	containerConfig := container.Config{
		Image:      opts.ImageTags,
		Cmd:        opts.Cmd,
		WorkingDir: opts.WorkingDir,
		Tty:        true,
	}

	hostConfig := container.HostConfig{
		Binds: binds,
	}

	resp, err := cli.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, nil, opts.Name)
	if err != nil {
		return fmt.Errorf("Failed to create container %s", err)
	}

	containerID := resp.ID

	defer func() {
		if opts.Debug {
			log.Warnf("Container %s is not removed due to debug mode", containerID)
			return
		}

		log.Infof("Remove container...")
		err := cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
			RemoveVolumes: true,
		})
		if err != nil {
			log.Warnf("Failed to remove container: %s", err)
		}
	}()

	if err := cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("Failed to start container: %s", err)
	}

	if err := waitForContainer(cli, containerID, opts.ShowOutput); err != nil {
		return fmt.Errorf("Failed to run command on container: %s", err)
	}

	return nil
}
