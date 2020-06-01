package create

import (
	"fmt"
	"os"

	"github.com/tarantool/cartridge-cli/cli/common"

	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/cli/create/templates"
	"github.com/tarantool/cartridge-cli/cli/project"
)

// Run creates a project in projectCtx.Path
func Run(projectCtx *project.ProjectCtx) error {
	log.Infof("Creating an application %s...", projectCtx.Name)

	common.CheckRecommendedBinaries("git")

	// check context
	if err := checkCtx(projectCtx); err != nil {
		// TODO: format internal error
		panic(err)
	}

	// check that application doesn't exist
	if _, err := os.Stat(projectCtx.Path); err == nil {
		return fmt.Errorf("Application already exists in %s", projectCtx.Path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Unable to create application in %s: %s", projectCtx.Path, err)
	}

	if err := os.Mkdir(projectCtx.Path, 0755); err != nil {
		return fmt.Errorf("Failed to create application directory: %s", err)
	}

	if err := templates.Instantiate(projectCtx); err != nil {
		os.RemoveAll(projectCtx.Path)
		return fmt.Errorf("Failed to instantiate application template: %s", err)
	}

	log.Infof("Instantiated application files")

	if err := initGitRepo(projectCtx); err != nil {
		log.Warnf("Failed to initialize git repo: %s", err)
	} else {
		log.Infof("Initialized git repo")
	}

	log.Infof("Application %q created successfully", projectCtx.Name)

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
