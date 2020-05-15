package pack

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"

	build "github.com/tarantool/cartridge-cli/build_project"
	"github.com/tarantool/cartridge-cli/common"
	"github.com/tarantool/cartridge-cli/project"
)

const (
	fileReqPerms = 0444
	dirReqPerms  = 0555
)

func initDistributionDir(destPath string, projectCtx *project.ProjectCtx) error {
	log.Debugf("Create distribution dir: %s", destPath)
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("Failed to create distribution dir: %s", err)
	}

	log.Debugf("Copy application files to: %s", destPath)
	err := copy.Copy(projectCtx.Path, destPath, copy.Options{
		Skip: func(src string) bool {
			relPath, err := filepath.Rel(projectCtx.Path, src)
			if err != nil {
				panic(err)
			}

			return relPath == ".rocks" || strings.HasPrefix(relPath, ".rocks/")
		},
	})
	if err != nil {
		return fmt.Errorf("Failed to copy application files: %s", err)
	}

	log.Debugf("Cleanup distribution files")
	if err := cleanupDistrbutionFiles(destPath, projectCtx); err != nil {
		return fmt.Errorf("Failed to copy application files: %s", err)
	}

	log.Debugf("Check filemodes")
	if err := checkFilemodes(destPath); err != nil {
		return err
	}

	log.Debugf("Build application in %s", destPath)
	projectCtx.BuildDir = destPath
	if err := build.Run(projectCtx); err != nil {
		return err
	}

	return nil
}

func cleanupDistrbutionFiles(destPath string, projectCtx *project.ProjectCtx) error {
	if !common.GitIsInstalled() {
		log.Warnf("git not found. It is possible that some of the extra files " +
			"normally ignored are shipped to the resulting package. ")
	} else if !common.IsGitProject(destPath) {
		log.Warnf("Directory %s is not a git project. It is possible that some of the extra files "+
			"normally ignored are shipped to the resulting package. ",
			destPath)
	} else {
		log.Debugf("Running `git clean`")
		gitCleanCmd := exec.Command("git", "clean", "-f", "-d", "-X")
		if err := common.RunCommand(gitCleanCmd, destPath, projectCtx.Debug); err != nil {
			log.Warnf("Failed to run `git clean`")
		}

		log.Debugf("Running `git clean` for submodules")
		gitSubmodulesCleanCmd := exec.Command(
			"git", "submodule", "foreach", "--recursive", "git", "clean", "-f", "-d", "-X",
		)
		if err := common.RunCommand(gitSubmodulesCleanCmd, destPath, projectCtx.Debug); err != nil {
			log.Warnf("Failed to run `git clean` for submodules")
		}
	}

	log.Debugf("Remove `.git` directory")
	if err := os.RemoveAll(filepath.Join(destPath, ".git")); err != nil {
		return fmt.Errorf("Failed to remove .git directory", err)
	}

	return nil
}

func checkFilemodes(destPath string) error {
	if fileInfo, err := os.Stat(destPath); err != nil {
		return err
	} else if !fileInfo.IsDir() {
		if !common.HasPerm(fileInfo, fileReqPerms) {
			return fmt.Errorf("File %s has invalid mode: %o. "+
				"It should have read permissions for all", destPath, fileInfo.Mode())
		}
	} else {
		f, err := os.Open(destPath)
		if err != nil {
			return err
		}

		fileInfos, err := f.Readdir(-1)
		f.Close()
		if err != nil {
			return err
		}

		for _, fileInfo := range fileInfos {
			filePath := filepath.Join(destPath, fileInfo.Name())
			if err := checkFilemodes(filePath); err != nil {
				return err
			}
		}

	}

	return nil
}
