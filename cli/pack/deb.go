package pack

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/apex/log"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
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
func packDeb(ctx *context.Ctx) error {
	var err error

	if err := common.CheckRequiredBinaries("ar"); err != nil {
		return err
	}

	// app dir
	dataDirPath := filepath.Join(ctx.Pack.PackageFilesDir, dataDirName)
	appDirPath := filepath.Join(dataDirPath, ctx.Running.AppDir)
	if err := initAppDir(appDirPath, ctx); err != nil {
		return err
	}

	// systemd dir
	if err := initSystemdDir(dataDirPath, ctx); err != nil {
		return err
	}

	// tmpfiles dir
	if err := initTmpfilesDir(dataDirPath, ctx); err != nil {
		return err
	}

	//  data.tar.gz
	log.Debugf("Create data archive")
	dataArchivePath := filepath.Join(ctx.Pack.PackageFilesDir, dataArchiveName)
	err = common.WriteTgzArchive(dataDirPath, dataArchivePath)
	if err != nil {
		return err
	}

	// control dir
	controlDirPath := filepath.Join(ctx.Pack.PackageFilesDir, controlDirName)
	if err := initControlDir(controlDirPath, ctx); err != nil {
		return err
	}

	// control.tar.gz
	log.Debugf("Create deb control directory archive")
	controlArchivePath := filepath.Join(ctx.Pack.PackageFilesDir, controlArchiveName)
	err = common.WriteTgzArchive(controlDirPath, controlArchivePath)
	if err != nil {
		return err
	}

	// debian-binary
	log.Debugf("Create debian-binary file")
	if err = debianBinaryTemplate.Instantiate(ctx.Pack.PackageFilesDir, nil); err != nil {
		return fmt.Errorf("Failed to create debian-binary file: %s", err)
	}

	// create result archive
	log.Infof("Create result DEB package...")
	packDebCmd := exec.Command(
		"ar", "r",
		ctx.Pack.ResPackagePath,
		// the order matters
		debianBinaryFileName,
		controlArchivePath,
		dataArchivePath,
	)

	err = common.RunCommand(packDebCmd, ctx.Pack.PackageFilesDir, ctx.Cli.Verbose)
	if err != nil {
		return fmt.Errorf("Failed to pack DEB: %s", err)
	}

	log.Infof("Created result DEB package: %s", ctx.Pack.ResPackagePath)

	return nil
}

const (
	dataDirName    = "data"
	controlDirName = "control"

	dataArchiveName    = "data.tar.gz"
	controlArchiveName = "control.tar.gz"

	debianBinaryFileName = "debian-binary"
)
