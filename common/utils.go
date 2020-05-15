package common

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
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

// TarantoolVersion gets Tarantool version
func GetTarantoolVersion(tarantoolDir string) (string, error) {
	var err error

	tarantool := filepath.Join(tarantoolDir, "tarantool")
	versionCmd := exec.Command(tarantool, "--version")

	tarantoolVersion, err := GetOutput(versionCmd, nil)
	if err != nil {
		return "", err
	}

	r := regexp.MustCompile(`\d+\.\d+\.\d+-\d+-\w+`)
	tarantoolVersion = r.FindString(tarantoolVersion)

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

func GetHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

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

func GitIsInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func IsGitProject(path string) bool {
	fileInfo, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && fileInfo.IsDir()
}

func HasPerm(fileInfo os.FileInfo, perm os.FileMode) bool {
	return fileInfo.Mode()&perm == perm
}

func FileLinesScanner(file *os.File) *bufio.Scanner {
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	return scanner
}

func CreateTgzArchive(srcDirPath string, destFilePAth string) error {
	resPackageFile, err := os.Create(destFilePAth)
	if err != nil {
		return fmt.Errorf("Failed to create result file %s: %s", destFilePAth, err)
	}

	gzipWriter := gzip.NewWriter(resPackageFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	filepath.Walk(srcDirPath, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fileInfo.Mode().IsRegular() {
			return nil
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

		if err := writeFileToWriter(filePath, tarWriter); err != nil {
			return err
		}

		return nil
	})

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
