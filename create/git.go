package create

import (
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/common"
	"github.com/tarantool/cartridge-cli/project"
)

const (
	initialTagName   = "0.1.0"
	initialCommitMsg = "Initial commit"
)

func initGitRepo(projectCtx *project.ProjectCtx) error {
	// check that git is installed
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git not found. " +
			"You'll need to add the app to version control yourself later")
	}

	// init repo
	initCmd := exec.Command("git", "init")
	if err := common.RunCommand(initCmd, projectCtx.Path, false); err != nil {
		return fmt.Errorf("Failed to initialize git repo")
	}

	log.Debug("Initialized git repo")

	// add files to index
	addCmd := exec.Command("git", "add", "-A")
	if err := common.RunCommand(addCmd, projectCtx.Path, false); err != nil {
		return fmt.Errorf("Failed to add file to index")
	}

	log.Debug("Added files to index")

	// create initial commit
	commitCmd := exec.Command("git", "commit", "-m", initialCommitMsg)
	if err := common.RunCommand(commitCmd, projectCtx.Path, false); err != nil {
		return fmt.Errorf("Failed to create initial commit")
	}

	log.Debug("Created initial commit")

	// create initial tag
	tagCmd := exec.Command("git", "tag", initialTagName)
	if err := common.RunCommand(tagCmd, projectCtx.Path, false); err != nil {
		return fmt.Errorf("Failed to create initial commit")
	}

	log.Debugf("Created initial tag %s", initialTagName)

	return nil
}
