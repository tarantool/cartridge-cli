package templates

import (
	"fmt"

	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/templates"
)

const (
	CartridgeTemplateName = "cartridge"
)

var (
	knownTemplates = map[string]*templates.FileTreeTemplate{}
)

func init() {
	knownTemplates[CartridgeTemplateName] = templates.Combine(
		appFilesTemplate,
		buildFilesTemplate,
		configFilesTemplate,
		devFilesTemplate,
		testFilesTemplate,
	)
}

// Instantiate creates a file tree in a ctx.Project.Path according to ctx.Project.Template
// It applies ctx.Project to the template
func Instantiate(ctx *context.Ctx) error {
	projectTmpl, exists := knownTemplates[ctx.Project.Template]
	if !exists {
		return fmt.Errorf("Template %s does not exists", ctx.Project.Template)
	}

	if err := projectTmpl.Instantiate(ctx.Project.Path, ctx.Project); err != nil {
		return fmt.Errorf("Failed to instantiate %s template: %s", ctx.Project.Template, err)
	}

	return nil
}
