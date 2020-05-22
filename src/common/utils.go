package common

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	tarantoolVersionRegexp *regexp.Regexp
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	tarantoolVersionRegexp = regexp.MustCompile(`\d+\.\d+\.\d+-\d+-\w+`)
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

// GetTarantoolDir returns Tarantool executable directory
func GetTarantoolDir() (string, error) {
	var err error

	tarantool, err := exec.LookPath("tarantool")
	if err != nil {
		return "", fmt.Errorf("tarantool executable not found")
	}

	return filepath.Dir(tarantool), nil
}

// TarantoolIsEnterprise checks if Tarantool is Enterprise
func TarantoolIsEnterprise(tarantoolDir string) (bool, error) {
	var err error

	tarantool := filepath.Join(tarantoolDir, "tarantool")
	versionCmd := exec.Command(tarantool, "--version")

	tarantoolVersion, err := GetOutput(versionCmd, nil)
	if err != nil {
		return false, err
	}

	return strings.HasPrefix(tarantoolVersion, "Tarantool Enterprise"), nil
}

// GetTarantoolVersion gets Tarantool version
func GetTarantoolVersion(tarantoolDir string) (string, error) {
	var err error

	tarantool := filepath.Join(tarantoolDir, "tarantool")
	versionCmd := exec.Command(tarantool, "--version")

	tarantoolVersion, err := GetOutput(versionCmd, nil)
	if err != nil {
		return "", err
	}

	tarantoolVersion = tarantoolVersionRegexp.FindString(tarantoolVersion)

	if tarantoolVersion == "" {
		return "", fmt.Errorf("Failed to match Tarantool version")
	}

	return tarantoolVersion, nil
}

// IsExecOwner checks if specified file has owner execute permissions
func IsExecOwner(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	perm := fileInfo.Mode().Perm()
	return perm&0100 != 0, nil
}

// FindRockspec finds *.rockspec file in specified path
// If multiple files are found, it returns an error
func FindRockspec(path string) (string, error) {
	rockspecs, err := filepath.Glob(filepath.Join(path, "*.rockspec"))

	if err != nil {
		return "", fmt.Errorf("Failed to find rockspec: %s", err)
	}

	if len(rockspecs) > 1 {
		return "", fmt.Errorf("Found multiple rockspecs in %s", path)
	}

	if len(rockspecs) == 1 {
		return rockspecs[0], nil
	}

	return "", nil
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
	defer file.Close()

	if err != nil {
		return "", err
	}

	var fileContent string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fileContent += scanner.Text()
	}

	return fileContent, nil

}

// WriteTarArchive creates Tar archive of specified path
// using specified writer
func WriteTarArchive(srcDirPath string, compressWriter io.Writer) error {
	tarWriter := tar.NewWriter(compressWriter)
	defer tarWriter.Close()

	err := filepath.Walk(srcDirPath, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		tarHeader, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
		if err != nil {
			return err
		}

		tarHeader.Name, err = filepath.Rel(srcDirPath, filePath)
		if err != nil {
			return err
		}

		if err := tarWriter.WriteHeader(tarHeader); err != nil {
			return err
		}

		if fileInfo.Mode().IsRegular() {
			if err := writeFileToWriter(filePath, tarWriter); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
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

// WriteTgzArchive creates TGZ archive of specified path
func WriteTgzArchive(srcDirPath string, destFilePath string) error {
	destFile, err := os.Create(destFilePath)
	if err != nil {
		return fmt.Errorf("Failed to create result TGZ file %s: %s", destFilePath, err)
	}

	gzipWriter := gzip.NewWriter(destFile)
	defer gzipWriter.Close()

	err = WriteTarArchive(srcDirPath, gzipWriter)
	if err != nil {
		return err
	}

	return nil
}

func CompressGzip(srcFilePath string, destFilePath string) error {
	var err error

	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		return fmt.Errorf("Failed to create result GZIP file %s: %s", srcFilePath, err)
	}

	srcFileScanner := bufio.NewScanner(srcFile)

	destFile, err := os.Create(destFilePath)
	if err != nil {
		return fmt.Errorf("Failed to create result GZIP file %s: %s", destFilePath, err)
	}

	gzipWriter := gzip.NewWriter(destFile)
	defer gzipWriter.Close()

	for srcFileScanner.Scan() {
		gzipWriter.Write(srcFileScanner.Bytes())
	}

	return nil
}

// GetNextMajorVersion computes next major version for a given one
// for example, for 1.10.3 it's 2
func GetNextMajorVersion(version string) (string, error) {
	parts := strings.SplitN(version, ".", 2)
	major, err := strconv.Atoi(parts[0])

	if err != nil {
		return "", fmt.Errorf("Failed to convert major to int: %s", err)
	}

	return strconv.Itoa(major + 1), nil
}

func FileSHA256Hex(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func FileMD5Hex(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := md5.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
