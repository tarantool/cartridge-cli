package pack

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"

	build "github.com/tarantool/cartridge-cli/build_project"
	"github.com/tarantool/cartridge-cli/common"
	"github.com/tarantool/cartridge-cli/project"
)

const (
	fileReqPerms    = 0444
	dirReqPerms     = 0555
	versionFileName = "VERSION"
)

type rocksVersionsMapType = map[string]string

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

	// build
	projectCtx.BuildDir = destPath
	if err := build.Run(projectCtx); err != nil {
		return err
	}

	// post-build
	if err := build.PostRun(projectCtx); err != nil {
		return err
	}

	// generate VERSION file
	if err := generateVersionFile(projectCtx); err != nil {
		log.Warnf("Failed to generate VERSION file: %s", err)
	}

	if projectCtx.TarantoolIsEnterprise && !projectCtx.BuildInDocker {
		if err := copyTarantoolBinaries(projectCtx); err != nil {
			return err
		}
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

func generateVersionFile(projectCtx *project.ProjectCtx) error {
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
	rocksVersionsMap, err := getRocksVersions(projectCtx)
	if err != nil {
		log.Warnf("can't process rocks manifest file. Dependency information can't be "+
			"shipped to the resulting package: %s", err)
	} else {
		for rockName, rockVersion := range rocksVersionsMap {
			if rockName != projectCtx.Name {
				rockLine := fmt.Sprintf("%s=%s", rockName, rockVersion)
				versionFileLines = append(versionFileLines, rockLine)
			}
		}
	}

	versionFilePath := filepath.Join(projectCtx.BuildDir, versionFileName)
	versionFile, err := os.Create(versionFilePath)
	if err != nil {
		return fmt.Errorf("Failed to write VERSION file %s: %s", versionFilePath, err)
	}

	defer versionFile.Close()

	versionFile.WriteString(strings.Join(versionFileLines, "\n") + "\n")

	return nil
}

func getRocksVersions(projectCtx *project.ProjectCtx) (rocksVersionsMapType, error) {
	rocksVersionsMap := rocksVersionsMapType{}

	manifestFilePath := filepath.Join(projectCtx.BuildDir, ".rocks/share/tarantool/rocks/manifest")
	if _, err := os.Stat(manifestFilePath); err == nil {
		L := lua.NewState()
		defer L.Close()

		if err := L.DoFile(manifestFilePath); err != nil {
			return nil, fmt.Errorf("Failed to read manifest file %s: %s", manifestFilePath, err)
		}

		depsL := L.Env.RawGetString("dependencies")
		depsLTable, ok := depsL.(*lua.LTable)
		if !ok {
			return nil, fmt.Errorf("Failed to read manifest file: dependencies is not a table")
		}

		depsLTable.ForEach(func(depNameL lua.LValue, depInfoL lua.LValue) {
			depName := depNameL.String()

			depInfoLTable, ok := depInfoL.(*lua.LTable)
			if !ok {
				log.Warnf("Failed to get %s dependency info", depName)
			} else {
				depInfoLTable.ForEach(func(depVersionL lua.LValue, _ lua.LValue) {
					depVersion := depVersionL.String()
					if _, found := rocksVersionsMap[depName]; found {
						log.Warnf(
							"Found multiple versions for %s dependency in rocks manifest",
							depName,
						)
					}
					rocksVersionsMap[depName] = depVersion
				})
			}
		})

	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("Failed to read manifest file %s: %s", manifestFilePath, err)
	}

	return rocksVersionsMap, nil
}

func copyTarantoolBinaries(projectCtx *project.ProjectCtx) error {
	if !projectCtx.TarantoolIsEnterprise {
		panic("Tarantool should be Enterprise")
	}

	log.Infof("Copy Tarantool Enterprise binaries")

	tarantoolBinaries := []string{
		"tarantool",
		"tarantoolctl",
	}

	for _, binary := range tarantoolBinaries {
		binaryPath := filepath.Join(projectCtx.TarantoolDir, binary)
		copiedBinaryPath := filepath.Join(projectCtx.BuildDir, binary)

		if err := copy.Copy(binaryPath, copiedBinaryPath); err != nil {
			return fmt.Errorf("Failed to copy %s binary: %s", binary, err)
		}
	}

	return nil
}
