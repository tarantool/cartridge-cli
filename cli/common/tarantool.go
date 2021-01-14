package common

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	goVersion "github.com/hashicorp/go-version"
	"github.com/tarantool/cartridge-cli/cli/connector"
)

var (
	tarantoolVersionRegexp *regexp.Regexp
)

func init() {
	tarantoolVersionRegexp = regexp.MustCompile(`\d+\.\d+\.\d+-\d+-\w+`)
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

// GetMajorMinorVersion computes returns `<major>.<minor>` string
// for a given version
func GetMajorMinorVersion(version string) string {
	parts := strings.SplitN(version, ".", 3)
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

func GetCartridgeVersionStr(conn *connector.Conn) (string, error) {
	req := connector.EvalReq(getCartridgeVersionBody).SetReadTimeout(3 * time.Second)

	var versionStrSlice []string
	if err := conn.ExecTyped(req, &versionStrSlice); err != nil {
		return "", fmt.Errorf("Failed to eval get Cartridge version function: %s", err)
	}

	if len(versionStrSlice) != 1 {
		return "", fmt.Errorf("Cartridge version received in a wrong format")
	}

	versionStr := versionStrSlice[0]

	return versionStr, nil
}

func GetMajorCartridgeVersion(conn *connector.Conn) (int, error) {
	cartridgeVersionStr, err := GetCartridgeVersionStr(conn)
	if err != nil {
		return 0, err
	}

	if cartridgeVersionStr == "scm-1" {
		return 2, nil
	}

	cartridgeVersion, err := goVersion.NewSemver(cartridgeVersionStr)
	if err != nil {
		return 0, fmt.Errorf("Failed to parse Tarantool version: %s", err)
	}

	major := cartridgeVersion.Segments()[0]

	return major, nil
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

const (
	getCartridgeVersionBody = `return require('cartridge').VERSION or '1'`
)
