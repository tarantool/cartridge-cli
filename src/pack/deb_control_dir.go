package pack

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/tarantool/cartridge-cli/common"
	"github.com/tarantool/cartridge-cli/project"
	"github.com/tarantool/cartridge-cli/templates"
)

type debControlCtx struct {
	Name         string
	Version      string
	Maintainer   string
	Architecture string
	Depends      string
}

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
				Content: createUserScript,
			},
			{
				Path:    "postinst",
				Mode:    0755,
				Content: postinstScriptContent,
			},
		},
	}
)

func initControlDir(destDirPath string, projectCtx *project.ProjectCtx) error {
	log.Debugf("Create DEB control dir: %s", destDirPath)
	if err := os.MkdirAll(destDirPath, 0755); err != nil {
		return fmt.Errorf("Failed to create DEB control dir: %s", err)
	}

	ctx := debControlCtx{
		Name:         projectCtx.Name,
		Version:      projectCtx.VersionRelease,
		Maintainer:   defaultMaintainer,
		Architecture: defaultArch,
	}

	if !projectCtx.TarantoolIsEnterprise {
		minTarantoolVersion := projectCtx.TarantoolVersion
		maxTarantoolVersion, err := common.GetNextMajorVersion(minTarantoolVersion)
		if err != nil {
			panic(err)
		}

		ctx.Depends = fmt.Sprintf(
			"tarantool (>= %s), tarantool (<< %s)",
			minTarantoolVersion,
			maxTarantoolVersion,
		)
	}

	if err := templates.InstantiateTree(&debControlDirTemplate, destDirPath, ctx); err != nil {
		return fmt.Errorf("Failed to instantiate DEB control dir: %s", err)
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

	postinstScriptContent = `
/bin/sh -c 'chown -R root:root /usr/share/tarantool/{{ .Name }}'
/bin/sh -c 'chown root:root /etc/systemd/system/{{ .Name }}.service'
/bin/sh -c 'chown root:root /etc/systemd/system/{{ .Name }}@.service'
/bin/sh -c 'chown root:root /usr/lib/tmpfiles.d/{{ .Name }}.conf'`
)
