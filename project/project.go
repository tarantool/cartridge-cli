package project

import (
	"fmt"
	"os"

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

	Verbose               bool
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
