package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tarantool/cartridge-cli/templates"

	"github.com/tarantool/cartridge-cli/common"
	"github.com/tarantool/cartridge-cli/create"
	"github.com/tarantool/cartridge-cli/project"

	"github.com/spf13/cobra"
)

var (
	projectCtx project.ProjectCtx
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
		projectCtx.BasePath = cmd.Flags().Arg(0)

		err := normalizeCtx(&projectCtx)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		project.FillCtx(&projectCtx)

		err = create.CreateProject(projectCtx)
		if err != nil {
			fmt.Printf("Failed to create project: %s\n", err)
			os.Exit(1)
		}
	},
}

func normalizeCtx(projectCtx *project.ProjectCtx) error {
	var err error

	// set current directory as a default path
	if projectCtx.BasePath == "" {
		projectCtx.Path, err = os.Getwd()
		if err != nil {
			fmt.Println("Failed to get current directory")
			os.Exit(1)
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
