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

	"github.com/adam-hanna/arrayOperations"
	"github.com/mitchellh/mapstructure"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/vmihailenco/msgpack/v5"
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
	fileContent, err := GetFileContentBytes(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file: %s", err)
	}

	res := make(map[string]interface{})
	if err := yaml.Unmarshal(fileContent, res); err != nil {
		return nil, fmt.Errorf("Failed to parse %s: %s", path, err)
	}

	return res, nil
}

func readFromPos(f *os.File, pos int64, buf *[]byte) (int, error) {
	if _, err := f.Seek(pos, io.SeekStart); err != nil {
		return 0, fmt.Errorf("Failed to seek: %s", err)
	}

	n, err := f.Read(*buf)
	if err != nil {
		return n, fmt.Errorf("Failed to read: %s", err)
	}

	return n, nil
}

// GetLastNLinesBegin return the position of last lines begin
func GetLastNLinesBegin(filepath string, lines int) (int64, error) {
	if lines == 0 {
		return 0, nil
	}

	if lines < 0 {
		lines = -lines
	}

	f, err := os.Open(filepath)
	if err != nil {
		return 0, fmt.Errorf("Failed to open file: %s", err)
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

	if _, err := readFromPos(f, fileSize-1, &buf); err != nil {
		return 0, err
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
		}

		n, err := readFromPos(f, filePos, &buf)
		if err != nil {
			return 0, err
		}

		for i := n - 1; i >= 0; i-- {
			b := buf[i]

			if b == '\n' {
				newLinesN++
			}

			if newLinesN == lines+1 {
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

func GetLastNLines(filepath string, linesN int) ([]string, error) {
	lastNLinesBeginPos, err := GetLastNLinesBegin(filepath, linesN)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open file: %s", err)
	}

	if _, err := file.Seek(lastNLinesBeginPos, io.SeekStart); err != nil {
		return nil, fmt.Errorf("Failed to seek in file: %s", err)
	}

	lines := []string{}

	scanner := FileLinesScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, nil
}

func ConvertToStringsSlice(s interface{}) ([]string, error) {
	sliceRaw, err := ConvertToSlice(s)
	if err != nil {
		return nil, err
	}

	stringsSlice := make([]string, len(sliceRaw))
	for i, elem := range sliceRaw {
		stringElem, ok := elem.(string)
		if !ok {
			return nil, fmt.Errorf("Slice element %d isn't a string: %v", i, elem)
		}

		stringsSlice[i] = stringElem
	}

	return stringsSlice, nil
}

func ConvertToSlice(raw interface{}) ([]interface{}, error) {
	iterfacesSlice, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Should be a list, got %#v", raw)
	}

	return iterfacesSlice, nil
}

func DecodeMsgpackStruct(d *msgpack.Decoder, v interface{}) error {
	m, err := d.DecodeMap()
	if err != nil {
		return err
	}

	if err := mapstructure.Decode(m, v); err != nil {
		return err
	}

	return nil
}

func StringsSliceElemIndex(s []string, elem string) int {
	for i, sliceElem := range s {
		if sliceElem == elem {
			return i
		}
	}
	return -1
}

func RemoveFromStringSlice(s []string, i int) []string {
	return append(s[:i], s[i+1:]...)
}

func InsertInStringSlice(s []string, i int, elem string) []string {
	res := make([]string, len(s)+1)

	copy(res[0:], s[0:i])
	res[i] = elem
	copy(res[i+1:], s[i:])

	return res
}

func GetInstancesFromArgs(args []string, ctx *context.Ctx) ([]string, error) {
	foundInstances := make(map[string]struct{})
	var instances []string

	for _, instanceName := range args {
		if instanceName == ctx.Project.Name {
			return nil, fmt.Errorf(appNameSpecifiedError)
		}

		parts := strings.SplitN(instanceName, ".", 2)

		if len(parts) > 1 {
			return nil, fmt.Errorf(instanceIDSpecified)
		}

		if instanceName != "" {
			if _, found := foundInstances[instanceName]; found {
				return nil, fmt.Errorf("Duplicate instance name specified: %s", instanceName)
			}

			instances = append(instances, instanceName)
			foundInstances[instanceName] = struct{}{}
		}
	}

	return instances, nil
}

func GetStringSlicesDifference(s1, s2 []string) []string {
	uniqueStrings := arrayOperations.DifferenceString(s1, s2)
	return arrayOperations.IntersectString(s1, uniqueStrings)
}

func StringSliceContains(s []string, elem string) bool {
	for _, sliceElem := range s {
		if sliceElem == elem {
			return true
		}
	}

	return false
}

func StdinHasUnreadData() (bool, error) {
	stdinStat, err := os.Stdin.Stat()
	if err != nil {
		return false, err
	}
	return (stdinStat.Mode() & os.ModeCharDevice) == 0, nil
}

const (
	appNameSpecifiedError = "Application name is specified. " +
		"Please, specify instance name(s)"
	instanceIDSpecified = `[APP_NAME].INSTANCE_NAME is specified. ` +
		"Please, specify instance name(s)"
)
