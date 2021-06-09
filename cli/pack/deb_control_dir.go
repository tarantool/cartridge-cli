package pack

import (
	"fmt"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/templates"
)

var (
	debControlDirTemplate = templates.FileTreeTemplate{
		Dirs: []templates.DirTemplate{},
		Files: []templates.FileTemplate{
			{
				Path:    "control",
				Mode:    0644,
				Content: controlFileContent,
			},
			{
				Path:    "preinst",
				Mode:    0755,
				Content: project.PreInstScriptContent,
			},
			{
				Path:    "postinst",
				Mode:    0755,
				Content: project.PostInstScriptContent,
			},
		},
	}
)

func getDebRelation(relation string) string {
	if relation == ">" || relation == "<" {
		// Deb format uses >> and << instead of > and <
		return fmt.Sprintf("%s%s", relation, relation)
	} else if relation == "==" {
		return "="
	}

	return relation
}

func addDependenciesDeb(debControlCtx *map[string]interface{}, deps common.PackDependencies) {
	var depsList []string

	for _, dep := range deps {
		for _, r := range dep.Relations {
			depsList = append(depsList, fmt.Sprintf("%s (%s %s)", dep.Name, getDebRelation(r.Relation), r.Version))
		}

		if len(dep.Relations) == 0 {
			depsList = append(depsList, dep.Name)
		}
	}

	(*debControlCtx)["Depends"] = strings.Join(depsList, ", ")
}

func initControlDir(destDirPath string, ctx *context.Ctx) error {
	log.Debugf("Create DEB control directory")
	if err := os.MkdirAll(destDirPath, 0755); err != nil {
		return fmt.Errorf("Failed to create DEB control directory: %s", err)
	}

	debControlCtx := map[string]interface{}{
		"Name":         ctx.Project.Name,
		"Version":      ctx.Pack.VersionRelease,
		"Maintainer":   defaultMaintainer,
		"Architecture": defaultArch,
		"Depends":      "",
		"UserPreInst":  "",
		"UserPostInst": "",
	}

	addDependenciesDeb(&debControlCtx, ctx.Pack.Deps)

	parsedScript, err := common.ParseUserInstallScript(ctx.Pack.PreInstallScript)
	if err != nil {
		return fmt.Errorf("Failed to parse pre-install script: %s", err)
	}
	debControlCtx["UserPreInst"] = parsedScript

	parsedScript, err = common.ParseUserInstallScript(ctx.Pack.PostInstallScript)
	if err != nil {
		return fmt.Errorf("Failed to parse post-install script: %s", err)
	}
	debControlCtx["UserPostInst"] = parsedScript

	if err := debControlDirTemplate.Instantiate(destDirPath, debControlCtx); err != nil {
		return fmt.Errorf("Failed to instantiate DEB control directory: %s", err)
	}

	return nil
}

const (
	defaultMaintainer = "Tarantool Cartridge Developer"
	defaultArch       = "all"

	controlFileContent = `Package: {{ .Name }}
Version: {{ .Version }}
Maintainer: {{ .Maintainer }}
Architecture: {{ .Architecture }}
Description: Tarantool Cartridge app: {{ .Name }}
Depends: {{ .Depends }}

`
)
