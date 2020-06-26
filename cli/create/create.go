package create

import (
	"fmt"
	"os"

	"github.com/tarantool/cartridge-cli/cli/common"

	"github.com/apex/log"

	"github.com/tarantool/cartridge-cli/cli/create/templates"
	"github.com/tarantool/cartridge-cli/cli/project"
)

// Run creates a project in projectCtx.Path
func Run(projectCtx *project.ProjectCtx) error {
	common.CheckRecommendedBinaries("git")

	if err := checkCtx(projectCtx); err != nil {
		return project.InternalError("Create context check failed: %s", err)
	}

	// check that application doesn't exist
	if _, err := os.Stat(projectCtx.Path); err == nil {
		return fmt.Errorf("Application already exists in %s", projectCtx.Path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Unable to create application in %s: %s", projectCtx.Path, err)
	}

	log.Infof("Create application %s", projectCtx.Name)

	if err := os.Mkdir(projectCtx.Path, 0755); err != nil {
		return fmt.Errorf("Failed to create application directory: %s", err)
	}

	log.Infof("Generate application files")

	if err := templates.Instantiate(projectCtx); err != nil {
		os.RemoveAll(projectCtx.Path)
		return fmt.Errorf("Failed to instantiate application template: %s", err)
	}

	log.Infof("Initialize application git repository")
	if err := initGitRepo(projectCtx); err != nil {
		log.Warnf("Failed to initialize git repository: %s", err)
	}

	log.Infof("Application %q created successfully", projectCtx.Name)

	return nil
}

func FillCtx(projectCtx *project.ProjectCtx) error {
	projectCtx.StateboardName = project.GetStateboardName(projectCtx)

	return nil
}

func checkCtx(projectCtx *project.ProjectCtx) error {
	if projectCtx.Name == "" {
		return fmt.Errorf("Name is missed")
	}

	if projectCtx.StateboardName == "" {
		return fmt.Errorf("StateboardName is missed")
	}

	if projectCtx.Path == "" {
		return fmt.Errorf("Path is missed")
	}

	if projectCtx.Template == "" {
		return fmt.Errorf("Template is missed")
	}

	return nil
}
