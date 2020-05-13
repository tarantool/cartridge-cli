package commands

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/tarantool/cartridge-cli/common"
	"github.com/tarantool/cartridge-cli/create"
	"github.com/tarantool/cartridge-cli/project"
	"github.com/tarantool/cartridge-cli/templates"

	"github.com/spf13/cobra"
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
	setLogLevel()

	projectCtx.BasePath = cmd.Flags().Arg(0)

	if err := normalizeCtx(&projectCtx); err != nil {
		return err
	}

	project.FillCtx(&projectCtx)

	if err := create.CreateProject(projectCtx); err != nil {
		return err
	}

	return nil
}

func normalizeCtx(projectCtx *project.ProjectCtx) error {
	var err error

	// set current directory as a default path
	if projectCtx.BasePath == "" {
		projectCtx.Path, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	// check parent path
	projectCtx.BasePath, err = filepath.Abs(projectCtx.BasePath)
	if err != nil {
		return fmt.Errorf("Failed to normalize args: %s", err)
	}

	fileInfo, err := os.Stat(projectCtx.BasePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("Specified path %s does not exists", projectCtx.BasePath)
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("Specified path %s is not a directory", projectCtx.BasePath)
	}

	// prompt name if not specified
	if projectCtx.Name == "" {
		projectCtx.Name = common.Prompt("Enter project name", "myapp")
	}

	// set project path
	projectCtx.Path = filepath.Join(projectCtx.BasePath, projectCtx.Name)

	return nil
}
