package common

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apex/log"
	"github.com/otiai10/copy"
)

const (
	execOwnerPerm = 0100
)

// IsExecOwner checks if specified file has owner execute permissions
func IsExecOwner(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	perm := fileInfo.Mode().Perm()
	return perm&execOwnerPerm != 0, nil
}

// IsSocket checks if specified file is a socket
func IsSocket(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	perm := fileInfo.Mode()
	return perm&os.ModeSocket != 0, nil
}

// IsSubDir checks if directory is subdirectory of other
func IsSubDir(subdir string, dir string) (bool, error) {
	subdirAbs, err := filepath.Abs(subdir)
	if err != nil {
		return false, err
	}

	dirAbs, err := filepath.Abs(dir)
	if err != nil {
		return false, err
	}

	if dirAbs == subdirAbs {
		return true, nil
	}

	return strings.HasPrefix(subdirAbs, fmt.Sprintf("%s/", dirAbs)), nil
}

// ClearDir removes all files from specified directory
func ClearDir(dirPath string) error {
	files, err := filepath.Glob(filepath.Join(dirPath, "*"))
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			return err
		}
	}
	return nil
}

// HasPerm checks if specified file has permissions
func HasPerm(fileInfo os.FileInfo, perm os.FileMode) bool {
	return fileInfo.Mode()&perm == perm
}

// FileLinesScanner returns scanner for file
func FileLinesScanner(file *os.File) *bufio.Scanner {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	return scanner
}

// GetFileContent returns file content
func GetFileContent(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileContent, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(fileContent), nil

}

func writeFileToWriter(filePath string, writer io.Writer) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// copy file data into tar writer
	if _, err := io.Copy(writer, file); err != nil {
		return err
	}

	return nil
}

// MergeFiles creates a file that is a concatenation of srcFilePaths
func MergeFiles(destFilePath string, srcFilePaths ...string) error {
	destFile, err := os.Create(destFilePath)
	if err != nil {
		return fmt.Errorf("Failed to create result file %s: %s", destFilePath, err)
	}
	defer destFile.Close()

	for _, srcFilePath := range srcFilePaths {
		srcFile, err := os.Open(srcFilePath)
		if err != nil {
			return fmt.Errorf("Failed to open source file %s: %s", srcFilePath, err)
		}

		_, err = io.Copy(destFile, srcFile)
		srcFile.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func PrintFromStart(file *os.File) error {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("Failed to seek file begin: %s", err)
	}
	if _, err := io.Copy(os.Stdout, file); err != nil {
		log.Warnf("Failed to print file content: %s", err)
	}

	return nil
}

func ReplaceFileLinesByRe(filePath string, re *regexp.Regexp, repl string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Failed to open file: %s", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	tmpFile, err := ioutil.TempFile("", "replace-*")
	if err != nil {
		return fmt.Errorf("Failed to create tmp file: %s", err)
	}
	defer os.RemoveAll(tmpFile.Name())
	defer tmpFile.Close()

	for scanner.Scan() {
		line := scanner.Text()
		line = re.ReplaceAllString(line, repl)
		if _, err := io.WriteString(tmpFile, line+"\n"); err != nil {
			return err
		}
	}

	if err := copy.Copy(tmpFile.Name(), filePath, copy.Options{Sync: true}); err != nil {
		return fmt.Errorf("Failed to copy tmp file: %s", err)
	}

	return nil
}
