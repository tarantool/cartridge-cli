package commands

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/src/common"
	"github.com/tarantool/cartridge-cli/src/create"
	"github.com/tarantool/cartridge-cli/src/create/templates"
	"github.com/tarantool/cartridge-cli/src/project"
)

const (
	defaultTemplate = "cartridge"
)

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringVar(&projectCtx.Name, "name", "", "Application name")
	createCmd.Flags().StringVar(
		&projectCtx.Template,
		"template",
		templates.CartridgeTemplateName,
		"Application template",
	)
}

var createCmd = &cobra.Command{
	Use:   "create [PATH]",
	Short: "Create an application from the Cartridge template",
	Long:  "Create an application in the specified PATH (default \".\")",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := runCreateCommand(cmd, args)
		if err != nil {
			log.Fatalf(err.Error())
		}
	},
}

func runCreateCommand(cmd *cobra.Command, args []string) error {
	var err error

	// prompt name if not specified
	if projectCtx.Name == "" {
		projectCtx.Name = common.Prompt("Enter project name", "myapp")
	}

	// get project path
	basePath := cmd.Flags().Arg(0)
	projectCtx.Path, err = getNewProjectPath(basePath)
	if err != nil {
		return err
	}

	// fill context
	if err := project.FillCtx(&projectCtx); err != nil {
		return err
	}

	// create project
	if err := create.Run(&projectCtx); err != nil {
		return err
	}

	return nil
}

func getNewProjectPath(basePath string) (string, error) {
	var err error

	if basePath == "" {
		basePath, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	basePath, err = filepath.Abs(basePath)
	if err != nil {
		return "", fmt.Errorf("Failed to get absolute path: %s", err)
	}

	// check base path
	fileInfo, err := os.Stat(basePath)
	if err != nil {
		return "", fmt.Errorf("Unable to use specified path %s: %s", basePath, err)
	}

	if !fileInfo.IsDir() {
		return "", fmt.Errorf("Specified path %s is not a directory", basePath)
	}

	return filepath.Join(basePath, projectCtx.Name), nil
}
