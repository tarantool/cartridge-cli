package create

import (
	"fmt"
	"os"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"

	"github.com/apex/log"

	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/create/codegen/static"
	"github.com/tarantool/cartridge-cli/cli/create/templates"
)

// Run creates a project in ctx.Project.Path
func Run(ctx *context.Ctx) error {
	common.CheckRecommendedBinaries("git")

	if err := checkCtx(ctx); err != nil {
		return project.InternalError("Create context check failed: %s", err)
	}

	// check that application doesn't exist
	if _, err := os.Stat(ctx.Project.Path); err == nil {
		return fmt.Errorf("Application already exists in %s", ctx.Project.Path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Unable to create application in %s: %s", ctx.Project.Path, err)
	}

	log.Infof("Create application %s", ctx.Project.Name)

	if err := os.Mkdir(ctx.Project.Path, 0755); err != nil {
		return fmt.Errorf("Failed to create application directory: %s", err)
	}

	if ctx.Create.From == "" {
		switch ctx.Create.Template {
		case "cartridge":
			ctx.Create.TemplateFS = static.CartridgeTemplateFS
		default:
			return fmt.Errorf("Invalid template name: %s", ctx.Create.Template)
		}
	}

	log.Infof("Generate application files")

	if err := templates.Instantiate(ctx); err != nil {
		os.RemoveAll(ctx.Project.Path)
		return fmt.Errorf("Failed to instantiate application template: %s", err)
	}

	log.Infof("Initialize application git repository")
	if err := initGitRepo(ctx); err != nil {
		log.Warnf("Failed to initialize git repository: %s", err)
	}

	log.Infof("Application %q created successfully", ctx.Project.Name)

	return nil
}

func FillCtx(ctx *context.Ctx) error {
	ctx.Project.StateboardName = project.GetStateboardName(ctx)

	return nil
}

func checkCtx(ctx *context.Ctx) error {
	if ctx.Project.Name == "" {
		return fmt.Errorf("Name is missed")
	}

	if ctx.Project.StateboardName == "" {
		return fmt.Errorf("StateboardName is missed")
	}

	if ctx.Project.Path == "" {
		return fmt.Errorf("Path is missed")
	}

	if ctx.Create.Template == "" && ctx.Create.From == "" {
		return fmt.Errorf("Template name or path is missed")
	}

	return nil
}
