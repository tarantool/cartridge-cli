package rpm

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tarantool/cartridge-cli/src/project"
)

func packCpio(resFileName string, projectCtx *project.ProjectCtx) error {
	var files []string

	err := filepath.Walk(projectCtx.PackageFilesDir, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		filePath, err = filepath.Rel(projectCtx.PackageFilesDir, filePath)
		if err != nil {
			return err
		}

		if _, known := knownFiles[filePath]; !known {
			files = append(files, filePath)
		}

		return nil
	})

	if err != nil {
		return err
	}

	filesBuffer := bytes.Buffer{}
	filesBuffer.WriteString(strings.Join(files, "\n"))

	cpioFile, err := os.Create(resFileName)
	if err != nil {
		return err
	}
	defer cpioFile.Close()

	cpioFileWriter := bufio.NewWriter(cpioFile)
	defer cpioFileWriter.Flush()

	var stderrBuf bytes.Buffer

	cmd := exec.Command("cpio", "-o", "-H", "newc")
	cmd.Stdin = &filesBuffer
	cmd.Stdout = cpioFileWriter
	cmd.Stderr = &stderrBuf
	cmd.Dir = projectCtx.PackageFilesDir

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Failed to run \n%s\n\nStderr: %s", cmd.String(), stderrBuf.String())
	}

	return nil
}
