package create

import (
	"fmt"
	"os"

	"github.com/tarantool/cartridge-cli/templates"

	"github.com/tarantool/cartridge-cli/project"
)

func CreateProject(projectCtx project.ProjectCtx) error {
	fmt.Printf("Creating a project %s\n", projectCtx.Name)

	// check that application doesn't exist
	if _, err := os.Stat(projectCtx.Path); err == nil {
		return fmt.Errorf("Application already exists in %s", projectCtx.Path)
	}

	var err error

	err = os.Mkdir(projectCtx.Path, 0755)
	if err != nil {
		return fmt.Errorf("Failed to create application directory: %s", err)
	}

	err = templates.Instantiate(projectCtx)
	if err != nil {
		os.RemoveAll(projectCtx.Path)
		return fmt.Errorf("Failed to instantiate application template: %s", err)
	}

	return nil
}
