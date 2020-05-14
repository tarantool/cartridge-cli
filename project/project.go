package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"

	"github.com/tarantool/cartridge-cli/common"
)

const (
	defaultTemplate = "cartridge"
)

type ProjectCtx struct {
	Name           string
	StateboardName string
	Path           string
	Template       string

	Verbose bool
	Debug   bool
	Quiet   bool

	BuildID               string
	BuildDir              string
	PackageFilesDir       string
	BuildInDocker         bool
	TarantoolDir          string
	TarantoolIsEnterprise bool

	Version              string
	Release              string
	VersionRelease       string
	Suffix               string
	PackType             string
	UnitTemplatePath     string
	InstUnitTemplatePath string
}

// FillCtx fills project context
func FillCtx(projectCtx *ProjectCtx) error {
	var err error

	if projectCtx.Path == "" {
		projectCtx.Path, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("Failed to get current directory: %s", err)
		}
	}

	projectCtx.Path, err = filepath.Abs(projectCtx.Path)
	if err != nil {
		return fmt.Errorf("Failed to get absolute path for %s: %s", projectCtx.Path, err)
	}

	if projectCtx.Name == "" {
		if _, err := os.Stat(projectCtx.Path); err == nil {
			projectCtx.Name, err = detectName(projectCtx.Path)
			if err != nil {
				return fmt.Errorf("Failed to detect application name: %s", err)
			}
		}
	}

	projectCtx.StateboardName = fmt.Sprintf("%s-stateboard", projectCtx.Name)

	projectCtx.TarantoolDir, err = common.GetTarantoolDir()
	if err != nil {
		return fmt.Errorf("Failed to find Tarantool executable: %s", err)
	}

	projectCtx.TarantoolIsEnterprise, err = common.TarantoolIsEnterprise(projectCtx.TarantoolDir)
	if err != nil {
		return fmt.Errorf("Failed to check Tarantool version: %s", err)
	}

	return nil
}

// CheckTarantoolBinaries checks if all required binaries are installed
func CheckTarantoolBinaries() error {
	var requiredBinaries = []string{
		"tarantool",
		"tarantoolctl",
	}

	// check recommended binaries
	for _, binary := range requiredBinaries {
		if _, err := exec.LookPath(binary); err != nil {
			return fmt.Errorf("Missed %s binary", binary)
		}
	}

	return nil
}

func detectName(path string) (string, error) {
	var err error

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", fmt.Errorf("path %s does not exists", path)
	}

	rockspecPath, err := common.FindRockspec(path)
	if err != nil {
		return "", err
	} else if rockspecPath == "" {
		return "", fmt.Errorf("Application directory should contain rockspec")
	}

	L := lua.NewState()
	defer L.Close()

	if err := L.DoFile(rockspecPath); err != nil {
		return "", fmt.Errorf("Failed to read rockspec %s: %s", path, err)
	}

	packageLuaVal := L.Env.RawGetString("package")
	if packageLuaVal.Type() == lua.LTNil {
		return "", fmt.Errorf("Field 'package' is not set in rockspec %s", rockspecPath)
	}

	if packageLuaVal.Type() != lua.LTString {
		return "", fmt.Errorf("Field 'package' must be string in rockspec %s", rockspecPath)
	}

	return packageLuaVal.String(), nil
}
