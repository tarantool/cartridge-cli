// +build mage

package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	"github.com/otiai10/copy"
)

func downloadFile(url string, dest string) error {
	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("Failed to create dest file: %s", err)
	}
	defer destFile.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Failed to get: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Response status isn't OK: %s", resp.Status)
	}

	if _, err := io.Copy(destFile, resp.Body); err != nil {
		return fmt.Errorf("Failed to write dest file: %s", err)
	}

	return nil
}

func replaceFileLines(filePath string, re *regexp.Regexp, repl string) error {
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
