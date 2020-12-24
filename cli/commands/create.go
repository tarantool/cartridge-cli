package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/create"
	"github.com/tarantool/cartridge-cli/cli/create/codegen/static"
	"github.com/tarantool/cartridge-cli/cli/create/templates"
)

func init() {
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

	rootCmd.AddCommand(createCmd)

	// FLAGS
	configureFlags(createCmd)

	createCmd.Flags().StringVar(&ctx.Project.Name, "name", "", createNameUsage)
	createCmd.Flags().StringVar(&ctx.Create.From, "from", "", createFromUsage)
	createCmd.Flags().StringVar(&ctx.Create.Template, "template", "", templateUsage)
}

func runCreateCommand(cmd *cobra.Command, args []string) error {
	var err error

	// prompt name if not specified
	if ctx.Project.Name == "" {
		ctx.Project.Name = common.Prompt("Enter project name", "myapp")
	}

	// get project path
	basePath := cmd.Flags().Arg(0)
	ctx.Project.Path, err = getNewProjectPath(basePath)
	if err != nil {
		return err
	}

	if ctx.Create.Template != "" && ctx.Create.From != "" {
		return fmt.Errorf("You can specify only one of --from and --template options")
	}

	if ctx.Create.Template == "" && ctx.Create.From == "" {
		ctx.Create.Template = templates.CartridgeTemplateName
	}

	if ctx.Create.FileSystem == nil {
		ctx.Create.FileSystem = static.CartridgeData
	}

	// fill context
	if err := create.FillCtx(&ctx); err != nil {
		return err
	}

	// create project
	if err := create.Run(&ctx); err != nil {
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

	return filepath.Join(basePath, ctx.Project.Name), nil
}
