package common

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

// Prompt a value with given text and default value
func Prompt(text, defaultValue string) string {
	if defaultValue == "" {
		fmt.Printf("%s: ", text)
	} else {
		fmt.Printf("%s [%s]: ", text, defaultValue)
	}

	var value string
	fmt.Scanf("%s", &value)

	if value == "" {
		value = defaultValue
	}

	return value
}

// GetHomeDir returns current home directory
func GetHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

// RandomString generates random string length n
func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

// GitIsInstalled checks if git binary is in the PATH
func GitIsInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// IsGitProject checks if specified path is a git project
func IsGitProject(path string) bool {
	fileInfo, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && fileInfo.IsDir()
}

// ConcatBuffers appends sources content to dest
func ConcatBuffers(dest *bytes.Buffer, sources ...*bytes.Buffer) error {
	for _, src := range sources {
		if _, err := io.Copy(dest, src); err != nil {
			return err
		}
	}

	return nil
}

// GetCurrentUserID returns current user UID
func GetCurrentUserID() (string, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", err
	}
	return currentUser.Uid, nil
}

// OnlyOneIsTrue checks if one and only one of
// boolean values is true
func OnlyOneIsTrue(values ...bool) bool {
	trueValuesCount := 0

	for _, value := range values {
		if value {
			trueValuesCount++
			if trueValuesCount > 1 {
				return false
			}
		}
	}

	return trueValuesCount == 1

}

// TrimSince trims a string starts with a given substring.
// For example, TrimSince("a = 1 # comment", "#") is "a = 1 "
func TrimSince(s string, since string) string {
	index := strings.Index(s, since)
	if index == -1 {
		return s
	}

	return s[:index]
}
