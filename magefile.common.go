// +build mage

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
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
