package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/tarantool/cartridge-cli/src/common"
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

	PackID                string
	TmpDir                string
	PackageFilesDir       string
	BuildDir              string
	BuildInDocker         bool
	ResPackagePath        string
	TarantoolDir          string
	TarantoolVersion      string
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

	projectCtx.TarantoolVersion, err = common.GetTarantoolVersion(projectCtx.TarantoolDir)
	if err != nil {
		return fmt.Errorf("Failed to get Tarantool version: %s", err)
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

	name, err := common.LuaReadStringVar(rockspecPath, "package")
	if err != nil {
		return "", fmt.Errorf("Failed to read `package` field from rockspec: %s", err)
	}

	return name, nil
}
