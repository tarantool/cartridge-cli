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

	if !ctx.Tarantool.TarantoolIsEnterprise {
		minTarantoolVersion := ctx.Tarantool.TarantoolVersion
		maxTarantoolVersion, err := common.GetNextMajorVersion(minTarantoolVersion)
		if err != nil {
			return project.InternalError("Failed to get next Tarantool major version: %s", err)
		}

		addDependency(&debControlCtx, common.PackDependency{
			Name: "tarantool",
			Relations: []common.DepRelation{
				{
					Relation: ">=",
					Version:  minTarantoolVersion,
				},
				{
					Relation: "<<",
					Version:  maxTarantoolVersion,
				},
			},
		})
	}

	// Parse and add dependencies
	if len(ctx.Pack.Deps) != 0 {
		deps, err := common.ParseDependencies(ctx.Pack.Deps)
		if err != nil {
			return fmt.Errorf("Failed to parse dependencies file: %s", err)
		}

		for _, dependency := range deps.FormatDeb() {
			addDependency(&debControlCtx, dependency)
		}

		// cut last ', ' symbols created by addDependency function
		depString := fmt.Sprintf("%s", (debControlCtx)["Depends"])
		(debControlCtx)["Depends"] = depString[:len(depString)-2]
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
