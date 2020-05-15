package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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
	Quiet   bool

	TarantoolDir          string
	TarantoolIsEnterprise bool
	BuildInDocker         bool
	BuildDir              string
}

// FillCtx fills project context
func FillCtx(projectCtx *ProjectCtx) error {
	var err error

	if projectCtx.Path == "" {
		var err error

		projectCtx.Path, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("Failed to get current directory: %s", err)
		}
	}

	projectCtx.Path, err = filepath.Abs(projectCtx.Path)
	if err != nil {
		return fmt.Errorf("Failed to get absolute path for %s: %s", projectCtx.Path, err)
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
			return fmt.Errorf("missed %s binary", binary)
		}
	}

	return nil
}
