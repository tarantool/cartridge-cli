package common

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	goVersion "github.com/hashicorp/go-version"
	"github.com/robfig/config"
)

var (
	tarantoolVersionRegexp *regexp.Regexp
)

const (
	tarantoolExeName     = "tarantool"
	tarantoolVersionFlag = "--version"
	sdkVersionFileName   = "VERSION"
	sdkVersionOptName    = "TARANTOOL_SDK"
)

func init() {
	tarantoolVersionRegexp = regexp.MustCompile(`\d+\.\d+\.\d+-\d+-\w+`)
}

// GetTarantoolDir returns Tarantool executable directory
func GetTarantoolDir() (string, error) {
	var err error

	tarantool, err := exec.LookPath(tarantoolExeName)
	if err != nil {
		return "", fmt.Errorf("tarantool executable not found")
	}

	return filepath.Dir(tarantool), nil
}

// TarantoolIsEnterprise checks if Tarantool is Enterprise
func TarantoolIsEnterprise(tarantoolDir string) (bool, error) {
	tarantool := filepath.Join(tarantoolDir, tarantoolExeName)
	versionCmd := exec.Command(tarantool, tarantoolVersionFlag)

	tarantoolVersion, err := GetOutput(versionCmd, nil)
	if err != nil {
		return false, err
	}

	return strings.HasPrefix(tarantoolVersion, "Tarantool Enterprise"), nil
}

// GetSDKVersion gets SDK version from VERSION file placed
// in the Tarantool directory
func GetSDKVersion(tarantoolDir string) (string, error) {
	sdkVersionFilePath := filepath.Join(tarantoolDir, sdkVersionFileName)
	if _, err := os.Stat(sdkVersionFilePath); err != nil {
		return "", fmt.Errorf("Failed to use SDK version file: %s", err)
	}

	c, err := config.ReadDefault(sdkVersionFilePath)
	if err != nil {
		return "", fmt.Errorf("Failed to read SDK version file: %s", err)
	}

	sdkVersion, err := c.RawStringDefault(sdkVersionOptName)
	if err != nil {
		return "", fmt.Errorf(
			"Failed to get SDK version from %s file. %s is not specified",
			sdkVersionFileName, sdkVersionOptName,
		)
	}

	return sdkVersion, nil
}

// GetTarantoolVersion gets Tarantool version
func GetTarantoolVersion(tarantoolDir string) (string, error) {
	tarantool := filepath.Join(tarantoolDir, tarantoolExeName)
	versionCmd := exec.Command(tarantool, tarantoolVersionFlag)

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

// CheckTarantoolVersion checks if specified Tarantool version is
// a valid semver version
func CheckTarantoolVersion(versionStr string) error {
	_, err := goVersion.NewSemver(versionStr)
	if err != nil {
		return fmt.Errorf("Is not a valid semver version: %s", err)
	}

	return nil
}

// GetMajorMinorVersion computes returns `<major>.<minor>` string
// for a given version
func GetMajorMinorVersion(versionStr string) string {
	parts := strings.SplitN(versionStr, ".", 3)
	major := parts[0]
	minor := parts[1]

	majorMinorVersion := fmt.Sprintf("%s.%s", major, minor)

	return majorMinorVersion
}

// GetNextMajorVersion computes next major version for a given one.
// For example, for 1.10.3 it's 2
func GetNextMajorVersion(versionStr string) (string, error) {
	version, err := goVersion.NewSemver(versionStr)
	if err != nil {
		return "", fmt.Errorf("Failed to parse Tarantool version: %s", err)
	}

	major := version.Segments()[0]
	return strconv.Itoa(major + 1), nil
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
