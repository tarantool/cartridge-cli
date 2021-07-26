package docker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

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

	ShowOutput bool
}

var readerSize = 4096

func printBuildOutput(out io.Writer, body io.ReadCloser) error {
	rd := bufio.NewReaderSize(body, readerSize)
	var output map[string]interface{}
	buf := bytes.Buffer{}

Loop:
	for {
		buf.Reset()

		for {
			line, isPrefix, err := rd.ReadLine()
			if err == io.EOF {
				break Loop
			} else if err != nil {
				return fmt.Errorf("Failed to read docker build output: %s", err)
			}

			buf.Write(line)

			if !isPrefix {
				break
			}
		}

		if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
			return fmt.Errorf("Failed to unmarshal log: %s", err)
		}

		if stream, ok := output["stream"]; ok {
			if streamStr, ok := stream.(string); !ok {
				return fmt.Errorf("Received non-string stream: %s", stream)
			} else if _, err := io.Copy(out, strings.NewReader(streamStr)); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Output doesn't contain stream field: %s", buf.String())
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
	c := make(common.ReadyChan, 1)

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
		go common.StartCommandSpinner(c, &wg, "")
	}

	wg.Add(1)
	go func(buildErr *error) {
		defer wg.Done()
		defer common.SendReady(c)

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
	var err error

	cli, err := client.NewClientWithOpts()
	if err != nil {
		return err
	}

	ctx := context.Background()

	var tarReader io.Reader

	err = common.RunFunctionWithSpinner(func() error {
		tarReader, err = getTarDirReader(opts.BuildDir, opts.TmpDir)
		if err != nil {
			return err
		}

		return nil
	}, "Compressing build context...")

	if err != nil {
		return fmt.Errorf("Failed to compress build context: %s", err)
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

	if err := waitBuildOutput(resp, opts.ShowOutput); err != nil {
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
