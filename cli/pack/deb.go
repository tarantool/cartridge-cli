package pack

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/apex/log"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/templates"
)

var (
	debianBinaryTemplate = templates.FileTemplate{
		Path:    debianBinaryFileName,
		Mode:    0644,
		Content: "2.0\n",
	}
)

// DEB package is an ar archive that contains debian-binary, control.tar.gz and data.tar.gz files

// debian-binary  : contains format version string (2.0)
// data.tar.xz    : package files
// control.tar.xz : control files (control, preinst etc.)
func packDeb(projectCtx *project.ProjectCtx) error {
	var err error

	if err := common.CheckRequiredBinaries("ar"); err != nil {
		return err
	}

	// app dir
	dataDirPath := filepath.Join(projectCtx.PackageFilesDir, dataDirName)
	appDirPath := filepath.Join(dataDirPath, projectCtx.AppDir)
	if err := initAppDir(appDirPath, projectCtx); err != nil {
		return err
	}

	// systemd dir
	if err := initSystemdDir(dataDirPath, projectCtx); err != nil {
		return err
	}

	// tmpfiles dir
	if err := initTmpfilesDir(dataDirPath, projectCtx); err != nil {
		return err
	}

	//  data.tar.gz
	dataArchivePath := filepath.Join(projectCtx.PackageFilesDir, dataArchiveName)
	err = common.WriteTgzArchive(dataDirPath, dataArchivePath)
	if err != nil {
		return err
	}

	// control dir
	controlDirPath := filepath.Join(projectCtx.PackageFilesDir, controlDirName)
	if err := initControlDir(controlDirPath, projectCtx); err != nil {
		return err
	}

	// control.tar.gz
	controlArchivePath := filepath.Join(projectCtx.PackageFilesDir, controlArchiveName)
	err = common.WriteTgzArchive(controlDirPath, controlArchivePath)
	if err != nil {
		return err
	}

	// debian-binary
	err = debianBinaryTemplate.Instantiate(projectCtx.PackageFilesDir, nil)

	if err != nil {
		return fmt.Errorf("Failed to create debian-binary file: %s", err)
	}

	// create result archive
	log.Debugf("Create DEB package")
	packDebCmd := exec.Command(
		"ar", "r",
		projectCtx.ResPackagePath,
		// the order matters
		debianBinaryFileName,
		controlArchivePath,
		dataArchivePath,
	)

	err = common.RunCommand(packDebCmd, projectCtx.PackageFilesDir, !projectCtx.Quiet)
	if err != nil {
		return fmt.Errorf("Failed to pack DEB: %s", err)
	}

	log.Infof("Created result package: %s", projectCtx.ResPackagePath)

	return nil
}

const (
	dataDirName    = "data"
	controlDirName = "control"

	dataArchiveName    = "data.tar.gz"
	controlArchiveName = "control.tar.gz"

	debianBinaryFileName = "debian-binary"
)
