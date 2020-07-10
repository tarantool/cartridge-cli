package create

import (
	"fmt"
	"os/exec"

	"github.com/apex/log"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

const (
	initialTagName   = "0.1.0"
	initialCommitMsg = "Initial commit"
)

func initGitRepo(ctx *context.Ctx) error {
	// check that git is installed
	if !common.GitIsInstalled() {
		return fmt.Errorf("git not found. " +
			"You'll need to add the application to version control yourself later")
	}

	log.Debug("Initialize empty git repository")
	initCmd := exec.Command("git", "init")
	if err := common.RunCommand(initCmd, ctx.Project.Path, false); err != nil {
		return fmt.Errorf("Failed to initialize git repository")
	}

	log.Debug("Add files to git index")
	addCmd := exec.Command("git", "add", "-A")
	if err := common.RunCommand(addCmd, ctx.Project.Path, false); err != nil {
		return fmt.Errorf("Failed to add files to index")
	}

	log.Debug("Create initial commit")
	commitCmd := exec.Command("git", "commit", "-m", initialCommitMsg)
	if err := common.RunCommand(commitCmd, ctx.Project.Path, false); err != nil {
		return fmt.Errorf("Failed to create initial commit")
	}

	log.Debugf("Create initial tag %s", initialTagName)
	tagCmd := exec.Command("git", "tag", initialTagName)
	if err := common.RunCommand(tagCmd, ctx.Project.Path, false); err != nil {
		return fmt.Errorf("Failed to create initial tag")
	}

	return nil
}
