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

		// We need a clean map every time - so as not to litter with
		// unnecessary fields from the last call json.Unmarshal.
		output := make(map[string]interface{})
		if err := json.Unmarshal(buf.Bytes(), &output); err != nil {
			return fmt.Errorf("Failed to unmarshal log: %s", err)
		}

		outputStr := convertDockerOutputToString(output)
		if _, err := io.Copy(out, strings.NewReader(outputStr)); err != nil {
			return err
		}

		if errMsg, ok := output["error"]; ok {
			return fmt.Errorf("Build failed: %s", errMsg)
		}
	}

	return nil
}

func convertDockerOutputToString(outputMap map[string]interface{}) string {
	// The data format is a bit strange: either it is completely
	// stored in the "stream" field, or it needs to be collected
	// piece by piece from the rest of the map.
	if outputMap["stream"] != nil {
		return outputMap["stream"].(string)
	}

	// Any of these fields can be null.
	var output string
	if outputMap["id"] != nil {
		output = fmt.Sprintf("%s: ", outputMap["id"])
	}

	if outputMap["status"] != nil {
		output = fmt.Sprintf("%s%s ", output, outputMap["status"])
	}

	if outputMap["progress"] != nil {
		output = fmt.Sprintf("%s%s", output, outputMap["progress"])
	}

	return fmt.Sprintf("%s\n", output)
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

	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
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
