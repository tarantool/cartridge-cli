package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	client "docker.io/go-docker"
	"docker.io/go-docker/api/types"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
)

type BuildOpts struct {
	Tag        []string
	Dockerfile string
	CacheFrom  []string
	NoCache    bool

	BuildDir string
	TmpDir   string

	Quiet bool
}

func printBuildOutput(out io.Writer, body io.ReadCloser) error {
	rd := bufio.NewReader(body)
	var output map[string]interface{}

	for {
		line, _, err := rd.ReadLine()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if err := json.Unmarshal(line, &output); err != nil {
			return err
		}

		if stream, ok := output["stream"]; ok {
			if streamStr, ok := stream.(string); !ok {
				return fmt.Errorf("Received non-string stream: %s", stream)
			} else if _, err := io.Copy(out, strings.NewReader(streamStr)); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Failed to parse line: %s", line)
		}

		if errMsg, ok := output["error"]; ok {
			return fmt.Errorf("Build failed: %s", errMsg)
		}
	}

	return nil
}

func waitBuildOutput(resp types.ImageBuildResponse, showOutput bool) error {
	var err error

	var wg sync.WaitGroup
	c := make(chan struct{}, 1)

	var outputBuf *os.File
	var out io.Writer

	if showOutput {
		out = os.Stdout
	} else {
		if outputBuf, err = ioutil.TempFile("", "out"); err != nil {
			out = ioutil.Discard
			log.Warnf("Failed to create tmp file to store docker build output: %s", err)
		} else {
			out = outputBuf
			defer outputBuf.Close()
			defer os.Remove(outputBuf.Name())
		}

		wg.Add(1)
		go common.StartCommandSpinner(c, &wg)
	}

	wg.Add(1)
	go func(buildErr *error) {
		defer wg.Done()
		defer func() { c <- struct{}{} }() // say that command is complete

		if err := printBuildOutput(out, resp.Body); err != nil {
			*buildErr = err
			return
		}
	}(&err)

	wg.Wait()

	if err != nil {
		if outputBuf != nil {
			if err := common.PrintFromStart(outputBuf); err != nil {
				log.Warnf("Failed to show docker build output: %s", err)
			}
		}

		return err
	}

	return nil
}

func BuildImage(opts BuildOpts) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	ctx := context.Background()

	tarReader, err := getTarDirReader(opts.BuildDir, opts.TmpDir)
	if err != nil {
		return err
	}

	resp, err := cli.ImageBuild(ctx, tarReader, types.ImageBuildOptions{
		Tags:       opts.Tag,
		Dockerfile: opts.Dockerfile,
		NoCache:    opts.NoCache,
		CacheFrom:  opts.CacheFrom,
		Remove:     true,
	})

	if err != nil {
		return err
	}

	if err := waitBuildOutput(resp, !opts.Quiet); err != nil {
		return err
	}

	return nil
}

func getTarDirReader(dirPath string, tmpDir string) (io.Reader, error) {
	tarFileName := fmt.Sprintf("%s.tar", filepath.Base(dirPath))
	tarFilePath := filepath.Join(tmpDir, tarFileName)

	tarWriter, err := os.Create(tarFilePath)
	if err != nil {
		return nil, err
	}

	if err := common.WriteTarArchive(dirPath, tarWriter); err != nil {
		return nil, err
	}

	tarReader, err := os.Open(tarFilePath)
	if err != nil {
		return nil, err
	}

	return tarReader, nil
}
