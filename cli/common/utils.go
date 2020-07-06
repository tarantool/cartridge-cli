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
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

var bufSize int64 = 10000

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

// IntsToStrings converts int slice to strings slice
func IntsToStrings(numbers []int) []string {
	var res []string

	for _, num := range numbers {
		res = append(res, strconv.Itoa(num))
	}

	return res
}

// ParseYmlFile reads YAML file and returns it's content as a map
func ParseYmlFile(path string) (map[string]interface{}, error) {
	fileContent, err := GetFileContent(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file: %s", err)
	}

	res := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(fileContent), res); err != nil {
		return nil, fmt.Errorf("Failed to parse %s: %s", path, err)
	}

	return res, nil
}

func readFromPos(f *os.File, pos int64, buf *[]byte) error {
	if _, err := f.Seek(pos, io.SeekStart); err != nil {
		return fmt.Errorf("Failed to seek: %s", err)
	}

	if _, err := f.Read(*buf); err != nil {
		return fmt.Errorf("Failed to read: %s", err)
	}

	return nil
}

// GetLastNLinesBegin return the position of last n lines begin
func GetLastNLinesBegin(filepath string, n int) (int64, error) {
	if n == 0 {
		return 0, nil
	}

	f, err := os.Open(filepath)
	if err != nil {
		return 0, fmt.Errorf("Failed to open log file: %s", err)
	}
	defer f.Close()

	var fileSize int64
	if fileInfo, err := os.Stat(filepath); err != nil {
		return 0, fmt.Errorf("Failed to get fileinfo: %s", err)
	} else {
		fileSize = fileInfo.Size()
	}

	if fileSize == 0 {
		return 0, nil
	}

	buf := make([]byte, bufSize)

	var filePos int64 = fileSize - bufSize
	var lastNewLinePos int64 = 0
	var newLinesN int = 0

	// check last symbol of the last line

	if err := readFromPos(f, fileSize-1, &buf); err != nil {
		return 0, fmt.Errorf("%s", err)
	}
	if buf[0] != '\n' {
		newLinesN++
	}

	lastPart := false

Loop:
	for {
		if filePos < 0 {
			filePos = 0
			lastPart = true

			buf = make([]byte, fileSize%bufSize)
		}

		if err := readFromPos(f, filePos, &buf); err != nil {
			return 0, fmt.Errorf("%s", err)
		}

		for i := len(buf) - 1; i >= 0; i-- {
			b := buf[i]

			if b == '\n' {
				newLinesN++
			}

			if newLinesN == n+1 {
				lastNewLinePos = filePos + int64(i+1)
				break Loop
			}
		}

		if lastPart || filePos == 0 {
			break
		}

		filePos -= bufSize
	}

	return lastNewLinePos, nil
}
