package pack

import (
	"fmt"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/tarantool/cartridge-cli/common"
	"github.com/tarantool/cartridge-cli/project"
)

func packTgz(projectCtx *project.ProjectCtx) error {
	if err := checkPackTgzRequiredBinaries(); err != nil {
		return err
	}

	// distribution dir
	distDir := filepath.Join(projectCtx.PackageFilesDir, projectCtx.Name)
	if err := initDistributionDir(distDir, projectCtx); err != nil {
		return err
	}

	// create archive
	err := common.CreateTgzArchive(projectCtx.PackageFilesDir, projectCtx.ResPackagePath)
	if err != nil {
		return err
	}

	log.Infof("Created result package: %s", projectCtx.ResPackagePath)

	return nil
}

func checkPackTgzRequiredBinaries() error {
	var requiredBinaries = []string{
		"tar",
	}

	for _, binary := range requiredBinaries {
		if _, err := exec.LookPath(binary); err != nil {
			return fmt.Errorf("%s binary is required to pack tgz", binary)
		}
	}

	return nil
}
