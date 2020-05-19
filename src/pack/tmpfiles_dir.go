package pack

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/tarantool/cartridge-cli/project"
	"github.com/tarantool/cartridge-cli/templates"
)

var (
	tmpFilesTemplate = templates.FileTreeTemplate{
		Dirs: []templates.DirTemplate{
			{
				Path: "/usr/lib/tmpfiles.d",
				Mode: 0755,
			},
		},
		Files: []templates.FileTemplate{
			{
				Path:    "/usr/lib/tmpfiles.d/{{ .Name }}.conf",
				Mode:    0644,
				Content: tmpFilesConfContent,
			},
		},
	}
)

func initTmpfilesDir(baseDirPath string, projectCtx *project.ProjectCtx) error {
	log.Debugf("Create tmpfiles dir in %s", baseDirPath)

	if err := tmpFilesTemplate.Instantiate(baseDirPath, projectCtx); err != nil {
		return fmt.Errorf("Failed to instantiate tmpfiles dir: %s", err)
	}

	return nil
}

const (
	tmpFilesConfContent = `d /var/run/tarantool 0755 tarantool tarantool`
)
