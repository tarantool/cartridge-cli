package pack

import (
	"fmt"
	"os"

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

func addDependency(debControlCtx *map[string]interface{}, deps common.PackDependency) {
	var depsString string

	for _, r := range deps.Relations {
		depsString = fmt.Sprintf("%s (%s %s)", deps.Name, r.Relation, r.Version)
		(*debControlCtx)["Depends"] = fmt.Sprintf("%s%s, ", (*debControlCtx)["Depends"], depsString)
	}

	if len(deps.Relations) == 0 {
		(*debControlCtx)["Depends"] = fmt.Sprintf("%s%s, ", (*debControlCtx)["Depends"], deps.Name)
	}
}

func formatDeb(deps common.PackDependencies) common.PackDependencies {
	debDeps := make(common.PackDependencies, 0, len(deps))

	for _, dependency := range deps {
		for i, r := range dependency.Relations {
			if r.Relation == ">" || r.Relation == "<" {
				// Deb format uses >> and << instead of > and <
				dependency.Relations[i].Relation = fmt.Sprintf("%s%s", r.Relation, r.Relation)
			} else if r.Relation == "==" {
				dependency.Relations[i].Relation = "="
			}
		}

		debDeps = append(debDeps, dependency)
	}

	return debDeps
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
	}

	var deps common.PackDependencies
	if !ctx.Tarantool.TarantoolIsEnterprise {
		var err error
		if deps, err = ctx.Pack.Deps.AddTarantool(ctx.Tarantool.TarantoolVersion); err != nil {
			return fmt.Errorf("Failed to add tarantool dependency: %s", err)
		}
	}

	deps = append(deps, ctx.Pack.Deps...)
	for _, dependency := range formatDeb(deps) {
		addDependency(&debControlCtx, dependency)
	}

	// cut last ', ' symbols created by addDependency function
	if len(ctx.Pack.Deps) != 0 {
		if depString := fmt.Sprintf("%s", (debControlCtx)["Depends"]); len(depString) != 0 {
			(debControlCtx)["Depends"] = depString[:len(depString)-2]
		}
	}

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
