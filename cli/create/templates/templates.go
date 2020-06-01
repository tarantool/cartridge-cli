package templates

import (
	"fmt"

	"github.com/tarantool/cartridge-cli/cli/project"
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

// Instantiate creates a file tree in a projectCtx.Path according to projectCtx.Template
// It applies ProjectCtx to the template
func Instantiate(projectCtx *project.ProjectCtx) error {
	projectTmpl, exists := knownTemplates[projectCtx.Template]
	if !exists {
		return fmt.Errorf("Template %s does not exists", projectCtx.Template)
	}

	if err := projectTmpl.Instantiate(projectCtx.Path, projectCtx); err != nil {
		return fmt.Errorf("Failed to instantiate %s template: %s", projectCtx.Template, err)
	}

	return nil
}
