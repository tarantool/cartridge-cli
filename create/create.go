package create

import (
	"fmt"
	"os"
	"os/exec"

	git "github.com/libgit2/git2go"

	"github.com/tarantool/cartridge-cli/project"
	"github.com/tarantool/cartridge-cli/templates"
)

const (
	initialTagName   = "0.1.0"
	initialCommitMsg = "Initial commit"
)

func CreateProject(projectCtx project.ProjectCtx) error {
	fmt.Printf("Creating a project %s...\n", projectCtx.Name)

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

	err = initGitRepo(projectCtx)
	if err != nil {
		fmt.Println(fmt.Errorf("Failed to initialize git repo: %s", err))
	}

	return nil
}

func initGitRepo(projectCtx project.ProjectCtx) error {
	var err error

	// check that git is installed
	if _, err = exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is not installed")
	}

	// init repo
	repo, err := git.InitRepository(projectCtx.Path, false)
	if err != nil {
		return fmt.Errorf("failed to init repo: %s", err)
	}

	fmt.Println("Initialized git repo")

	// get default git user
	userSignature, err := repo.DefaultSignature()
	if err != nil {
		return fmt.Errorf(`failed to get default git user: %s.
Please, call

    git config --global user.name "Your Name"
    git config --global user.email you@example.com

to set user`,
			err)
	}

	// add files to index
	index, err := repo.Index()
	if err != nil {
		return err
	}

	err = index.AddAll([]string{"."}, git.IndexAddDefault, nil)
	if err != nil {
		return fmt.Errorf("failed to add files to index: %s", err)
	}

	err = index.Write()
	if err != nil {
		return fmt.Errorf("failed to add files to index: %s", err)
	}

	fmt.Println("Application files are added to repo")

	// create initial commit
	oid, err := index.WriteTree()
	if err != nil {
		return fmt.Errorf("failed to create initial commit: %s", err)
	}

	commitID, err := repo.CreateCommitFromIds(
		"HEAD",
		userSignature,
		userSignature,
		initialCommitMsg,
		oid,
	)
	if err != nil {
		return fmt.Errorf("failed to create initial commit: %s", err)
	}

	fmt.Println("Initial commit is created")

	// create initial tag
	commit, err := repo.LookupCommit(commitID)
	if err != nil {
		return fmt.Errorf("failed to create initial tag: %s", err)
	}

	_, err = repo.Tags.CreateLightweight(
		initialTagName,
		commit,
		false,
	)
	if err != nil {
		return err
	}

	fmt.Println("Initial tag is created")

	return nil
}
