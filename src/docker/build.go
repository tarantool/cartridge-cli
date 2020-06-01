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
	"sync"

	"docker.io/go-docker/api/types"
	client "docker.io/go-docker"

	"github.com/tarantool/cartridge-cli/src/common"
)

type BuildOpts struct {
	Tag        string
	Dockerfile string
	CacheFrom  []string

	BuildDir string
	TmpDir   string

	Quiet bool
}

func printBuildOutput(body io.ReadCloser) error {
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
		stream, ok := output["stream"]
		if !ok {
			return fmt.Errorf("Output hasn't field `stream`: %s", string(line))
		}
		fmt.Printf("%s", stream)
	}

	return nil
}

func waitBuildOutput(resp types.ImageBuildResponse, quiet bool) error {
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

			if _, *err = io.Copy(ioutil.Discard, resp.Body); err != nil {
				return
			}

		}(&err)

		wg.Wait()
	} else {
		printBuildOutput(resp.Body)
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
		Tags:       []string{opts.Tag},
		Dockerfile: opts.Dockerfile,
		Remove:     true,
	})

	if err != nil {
		return err
	}

	if err := waitBuildOutput(resp, opts.Quiet); err != nil {
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
