package pack

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/otiai10/copy"

	"github.com/tarantool/cartridge-cli/cli/build"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"
)

const (
	fileReqPerms    = 0444
	dirReqPerms     = 0555
	versionFileName = "VERSION"
)

func initAppDir(appDirPath string, projectCtx *project.ProjectCtx) error {
	var err error

	log.Infof("Create application dir: %s", appDirPath)
	if err := os.MkdirAll(appDirPath, 0755); err != nil {
		return fmt.Errorf("Failed to create application dir: %s", err)
	}

	err = common.RunFunctionWithSpinner(func() error {
		err := copyProjectFiles(appDirPath, projectCtx)
		return err
	}, "Copying application files...")

	if err != nil {
		return fmt.Errorf("Failed to copy application files: %s", err)
	}

	log.Debugf("Cleanup application files")
	if err := cleanupAppDir(appDirPath, projectCtx); err != nil {
		return fmt.Errorf("Failed to copy application files: %s", err)
	}

	log.Debugf("Check filemodes")
	if err := checkFilemodes(appDirPath); err != nil {
		return err
	}

	// build
	projectCtx.BuildDir = appDirPath
	if err := build.Run(projectCtx); err != nil {
		return err
	}

	// post-build
	if err := build.PostRun(projectCtx); err != nil {
		return err
	}

	// generate VERSION file
	if err := generateVersionFile(appDirPath, projectCtx); err != nil {
		log.Warnf("Failed to generate VERSION file: %s", err)
	}

	if projectCtx.TarantoolIsEnterprise {
		log.Debugf("Copy Tarantool binaries")
		// copy Tarantool binaries to BuildDir to deliver in the result package
		if err := copyTarantoolBinaries(projectCtx.SDKPath, appDirPath); err != nil {
			return err
		}
	}

	return nil
}

func copyProjectFiles(dst string, projectCtx *project.ProjectCtx) error {
	err := copy.Copy(projectCtx.Path, dst, copy.Options{
		Skip: func(src string) bool {
			if strings.HasPrefix(src, fmt.Sprintf("%s/", projectCtx.CartridgeTmpDir)) {
				return true
			}

			relPath, err := filepath.Rel(projectCtx.Path, src)
			if err != nil {
				panic(err)
			}

			if relPath == ".rocks" || strings.HasPrefix(relPath, ".rocks/") {
				return true
			}

			if _, err := os.Open(src); err != nil {
				log.Warnf("Failed to copy: %s", err)
				return true
			}

			return false
		},
	})

	if err != nil {
		return fmt.Errorf("Failed to copy: %s", err)
	}

	return nil
}

func cleanupAppDir(appDirPath string, projectCtx *project.ProjectCtx) error {
	if !common.GitIsInstalled() {
		log.Warnf("git not found. It is possible that some of the extra files " +
			"normally ignored are shipped to the resulting package. ")
	} else if !common.IsGitProject(appDirPath) {
		log.Warnf("Directory %s is not a git project. It is possible that some of the extra files "+
			"normally ignored are shipped to the resulting package. ",
			appDirPath)
	} else {
		log.Debugf("Running `git clean`")
		gitCleanCmd := exec.Command("git", "clean", "-f", "-d", "-X")
		if err := common.RunCommand(gitCleanCmd, appDirPath, projectCtx.Debug); err != nil {
			log.Warnf("Failed to run `git clean`")
		}

		log.Debugf("Running `git clean` for submodules")
		gitSubmodulesCleanCmd := exec.Command(
			"git", "submodule", "foreach", "--recursive", "git", "clean", "-f", "-d", "-X",
		)
		if err := common.RunCommand(gitSubmodulesCleanCmd, appDirPath, projectCtx.Debug); err != nil {
			log.Warnf("Failed to run `git clean` for submodules")
		}
	}

	log.Debugf("Remove `.git` directory")
	if err := os.RemoveAll(filepath.Join(appDirPath, ".git")); err != nil {
		return fmt.Errorf("Failed to remove .git directory: %s", err)
	}

	return nil
}

func checkFilemodes(appDirPath string) error {
	if fileInfo, err := os.Stat(appDirPath); err != nil {
		return err
	} else if !fileInfo.IsDir() {
		if !common.HasPerm(fileInfo, fileReqPerms) {
			return fmt.Errorf("File %s has invalid mode: %o. "+
				"It should have read permissions for all", appDirPath, fileInfo.Mode())
		}
	} else {
		f, err := os.Open(appDirPath)
		if err != nil {
			return err
		}

		fileInfos, err := f.Readdir(-1)
		f.Close()
		if err != nil {
			return err
		}

		for _, fileInfo := range fileInfos {
			filePath := filepath.Join(appDirPath, fileInfo.Name())
			if err := checkFilemodes(filePath); err != nil {
				return err
			}
		}
	}

	return nil
}

func generateVersionFile(appDirPath string, projectCtx *project.ProjectCtx) error {
	log.Infof("Generate %s file", versionFileName)

	var versionFileLines []string

	// application version
	appVersionLine := fmt.Sprintf("%s=%s", projectCtx.Name, projectCtx.VersionRelease)
	versionFileLines = append(versionFileLines, appVersionLine)

	// Tarantool version
	if projectCtx.TarantoolIsEnterprise {
		tarantoolVersionFilePath := filepath.Join(projectCtx.TarantoolDir, "VERSION")
		tarantoolVersionFile, err := os.Open(tarantoolVersionFilePath)
		defer tarantoolVersionFile.Close()

		if err != nil {
			log.Warnf("Can't open VERSION file from Tarantool SDK: %s. SDK information can't be "+
				"shipped to the resulting package. ", err)
		}

		scanner := common.FileLinesScanner(tarantoolVersionFile)
		for scanner.Scan() {
			versionFileLines = append(versionFileLines, scanner.Text())
		}
	} else {
		tarantoolVersionLine := fmt.Sprintf("TARANTOOL=%s", projectCtx.TarantoolVersion)
		versionFileLines = append(versionFileLines, tarantoolVersionLine)
	}

	// rocks versions
	rocksVersionsMap, err := common.LuaGetRocksVersions(appDirPath)
	if err != nil {
		log.Warnf("Can't process rocks manifest file. Dependency information can't be "+
			"shipped to the resulting package: %s", err)
	} else {
		for rockName, rockVersion := range rocksVersionsMap {
			if rockName != projectCtx.Name {
				rockLine := fmt.Sprintf("%s=%s", rockName, rockVersion)
				versionFileLines = append(versionFileLines, rockLine)
			}
		}
	}

	versionFilePath := filepath.Join(appDirPath, versionFileName)
	versionFile, err := os.Create(versionFilePath)
	if err != nil {
		return fmt.Errorf("Failed to write VERSION file %s: %s", versionFilePath, err)
	}
	defer versionFile.Close()

	versionFile.WriteString(strings.Join(versionFileLines, "\n") + "\n")

	return nil
}

func copyTarantoolBinaries(binariesPath string, appDirPath string) error {
	tarantoolBinaries := []string{
		"tarantool",
		"tarantoolctl",
	}

	for _, binary := range tarantoolBinaries {
		binaryPath := filepath.Join(binariesPath, binary)
		destBinaryPath := filepath.Join(appDirPath, binary)

		if err := copy.Copy(binaryPath, destBinaryPath); err != nil {
			return fmt.Errorf("Failed to copy %s binary: %s", binary, err)
		}
	}

	return nil
}
