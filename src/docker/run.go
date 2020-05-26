package docker

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/tarantool/cartridge-cli/src/common"
)

type RunOpts struct {
	Name       string
	ImageTag   string
	WorkingDir string
	Cmd        []string

	Volumes map[string]string

	Quiet bool
	Debug bool
}

func waitStartOutput(out io.ReadCloser, quiet bool) error {
	if quiet {
		var err error

		var wg sync.WaitGroup
		c := make(chan struct{}, 1)

		wg.Add(1)
		go common.StartCommandSpinner(c, &wg)

		wg.Add(1)
		go func(err *error) {
			defer wg.Done()
			defer func() { c <- struct{}{} }() // say that command is complete

			if _, *err = io.Copy(ioutil.Discard, out); err != nil {
				return
			}

		}(&err)

		wg.Wait()
	} else {
		if _, err := io.Copy(os.Stdout, out); err != nil {
			return err
		}
	}
	return nil
}

func RunContainer(opts RunOpts) error {
	cli, err := client.NewEnvClient()
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
		Image:      opts.ImageTag,
		Cmd:        opts.Cmd,
		WorkingDir: opts.WorkingDir,
		Tty:        true,
	}

	hostConfig := container.HostConfig{
		Binds: binds,
	}

	// create container
	resp, err := cli.ContainerCreate(ctx, &containerConfig, &hostConfig, nil, opts.Name)
	if err != nil {
		return fmt.Errorf("Failed to create container %s", err)
	}

	containerID := resp.ID

	defer func() {
		if opts.Debug {
			log.Warnf("Build container %s is not removed due to debug mode", containerID)
			return
		}

		log.Infof("Remove build container...")
		err := cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
			// RemoveVolumes: true,
		})
		if err != nil {
			log.Warnf("Failed to remove build container: %s", err)
		}
	}()

	// start container
	if err := cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("Failed to start container: %s", err)
	}

	out, err := cli.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		Follow:     true,
	})
	if err != nil {
		return fmt.Errorf("Failed to check container logs: %s", err)
	}

	if err := waitStartOutput(out, opts.Quiet); err != nil {
		return err
	}

	if code, err := cli.ContainerWait(ctx, containerID); err != nil {
		return fmt.Errorf("Failed to run container: %s", err)
	} else if code != 0 {
		return fmt.Errorf("Failed to run command on container")
	}

	return nil
}
