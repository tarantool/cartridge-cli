package pack

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/apex/log"
	"github.com/otiai10/copy"

	"github.com/tarantool/cartridge-cli/cli/build"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

const (
	fileReqPerms       = 0444
	dirReqPerms        = 0555
	versionFileName    = "VERSION"
	versionLuaFileName = "VERSION.lua"
	maxCachedProjects  = 5
)

type cacheProjectPaths struct {
	projectPath string
	rocksPath   string
}

func initAppDir(appDirPath string, ctx *context.Ctx) error {
	var err error

	log.Infof("Initialize application dir")
	if err := os.MkdirAll(appDirPath, 0755); err != nil {
		return fmt.Errorf("Failed to create application dir: %s", err)
	}

	err = common.RunFunctionWithSpinner(func() error {
		err := copyProjectFiles(appDirPath, ctx)
		return err
	}, "Copying application files...")

	if err != nil {
		return fmt.Errorf("Failed to copy application files: %s", err)
	}

	log.Debugf("Cleanup application files")
	if err := cleanupAppDir(appDirPath, ctx); err != nil {
		return fmt.Errorf("Failed to copy application files: %s", err)
	}

	log.Debugf("Check filemodes")
	if err := checkFilemodes(appDirPath); err != nil {
		return err
	}

	cacheProjectPaths, err := getProjectCachePaths(ctx)
	if err != nil {
		return err
	}

	// Getting rocks
	ctx.Build.Dir = appDirPath

	if _, err := os.Stat(cacheProjectPaths.rocksPath); err == nil {
		// If rocks found in cache - we just copy them.
		if err := copyRocksFromCache(cacheProjectPaths.rocksPath, appDirPath); err != nil {
			return err
		}
	}

	if err := build.Run(ctx); err != nil {
		return err
	}

	// Update rocks cache in cartridge temp directory
	if err := updateRocksCache(ctx, cacheProjectPaths); err != nil {
		return err
	}

	// post-build
	if err := build.PostRun(ctx); err != nil {
		return err
	}

	// generate VERSION file
	if err := generateVersionFile(appDirPath, ctx); err != nil {
		log.Warnf("Failed to generate VERSION file: %s", err)
	}

	// generate VERSION.lua file
	if err := generateVersionLuaFile(appDirPath, ctx); err != nil {
		log.Warnf("Failed to generate VERSION.lua file: %s", err)
	}

	if ctx.Tarantool.TarantoolIsEnterprise {
		log.Debugf("Copy Tarantool binaries")
		// copy Tarantool binaries to BuildDir to deliver in the result package
		if err := copyTarantoolBinaries(ctx.Build.SDKPath, appDirPath); err != nil {
			return err
		}
	}

	return nil
}

func copyRocksFromCache(cachedRocksDir string, appDirPath string) error {
	log.Info("Getting rocks from a cache")
	err := copy.Copy(cachedRocksDir, appDirPath)

	if err != nil {
		return fmt.Errorf("Failed to copy .rocks from cache to project directory: %s", err)
	}

	return nil
}

func getProjectCachePaths(ctx *context.Ctx) (*cacheProjectPaths, error) {
	rockspecPath, err := common.FindRockspec(ctx.Project.Path)
	if err != nil {
		return nil, fmt.Errorf("Unable to find rockspec: %s", err)
	} else if rockspecPath == "" {
		return nil, fmt.Errorf("Application directory should contain rockspec")
	}

	projectPathHash := common.StringSHA256Hex(ctx.Project.Path)
	rockspecHash, err := common.FileSHA256Hex(rockspecPath)
	if err != nil {
		return nil, err
	}

	return &cacheProjectPaths{
		projectPath: filepath.Join(ctx.Cli.CacheDir, projectPathHash),
		rocksPath:   filepath.Join(ctx.Cli.CacheDir, projectPathHash, rockspecHash)}, nil
}

func removeRocksByEditTime(ctx *context.Ctx) error {
	dir, err := ioutil.ReadDir(ctx.Cli.CacheDir)
	if err != nil {
		return err
	}

	if len(dir) > maxCachedProjects {
		// Sort project rocks by change time
		sort.Slice(dir, func(i, j int) bool {
			return dir[i].ModTime().Before(dir[j].ModTime())
		})

		for i := 0; i < len(dir)-maxCachedProjects; i++ {
			projectRocksPath := filepath.Join(ctx.Cli.CacheDir, dir[i].Name())
			log.Debugf("Removing project rocks: %s", projectRocksPath)
			os.RemoveAll(projectRocksPath)
		}
	}

	return nil
}

func updateRocksCache(ctx *context.Ctx, cachePaths *cacheProjectPaths) error {
	if _, err := os.Stat(cachePaths.projectPath); err == nil {
		dir, err := ioutil.ReadDir(cachePaths.projectPath)
		if err != nil {
			return err
		}

		for _, d := range dir {
			os.RemoveAll(filepath.Join(cachePaths.projectPath, d.Name()))
		}
	}

	if err := copy.Copy(filepath.Join(ctx.Build.Dir, ".rocks"), filepath.Join(cachePaths.rocksPath, ".rocks")); err != nil {
		return fmt.Errorf("Failed to copy: %s", err)
	}

	log.Debugf("Rocks cache has been successfully saved in: %s", cachePaths.rocksPath)
	return removeRocksByEditTime(ctx)
}

func copyProjectFiles(dst string, ctx *context.Ctx) error {
	err := copy.Copy(ctx.Project.Path, dst, copy.Options{
		Skip: func(src string) (bool, error) {
			if strings.HasPrefix(src, fmt.Sprintf("%s/", ctx.Cli.CartridgeTmpDir)) {
				return true, nil
			}

			relPath, err := filepath.Rel(ctx.Project.Path, src)
			if err != nil {
				return false, fmt.Errorf("Failed to get file rel path: %s", err)
			}

			if relPath == ".rocks" || strings.HasPrefix(relPath, ".rocks/") {
				return true, nil
			}

			if isSocket, err := common.IsSocket(src); err != nil {
				return false, fmt.Errorf("Failed to check if file is a socket: %s", src)
			} else if isSocket {
				return true, nil
			}

			return false, nil
		},
	})

	if err != nil {
		return fmt.Errorf("Failed to copy: %s", err)
	}

	return nil
}

func cleanupAppDir(appDirPath string, ctx *context.Ctx) error {
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
		if err := common.RunCommand(gitCleanCmd, appDirPath, ctx.Cli.Debug); err != nil {
			log.Warnf("Failed to run `git clean`")
		}

		log.Debugf("Running `git clean` for submodules")
		gitSubmodulesCleanCmd := exec.Command(
			"git", "submodule", "foreach", "--recursive", "git", "clean", "-f", "-d", "-X",
		)
		if err := common.RunCommand(gitSubmodulesCleanCmd, appDirPath, ctx.Cli.Debug); err != nil {
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

func generateVersionLuaFile(appDirPath string, ctx *context.Ctx) error {
	log.Infof("Generate %s file", versionLuaFileName)

	versionLuaFilePath := filepath.Join(appDirPath, versionLuaFileName)
	// Check if the file already exists
	if _, err := os.Stat(versionLuaFilePath); err == nil {
		log.Warnf("File %s will be overwritten", versionLuaFileName)
	}

	versionLuaFile, err := os.OpenFile(versionLuaFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("Failed to write VERSION.lua file %s: %s", versionLuaFilePath, err)
	}
	defer versionLuaFile.Close()

	versionLuaFile.WriteString(fmt.Sprintf("return '%s'", ctx.Pack.VersionRelease))

	return nil
}

func generateVersionFile(appDirPath string, ctx *context.Ctx) error {
	log.Infof("Generate %s file", versionFileName)

	var versionFileLines []string

	// application version
	appVersionLine := fmt.Sprintf("%s=%s", ctx.Project.Name, ctx.Pack.VersionRelease)
	versionFileLines = append(versionFileLines, appVersionLine)

	// Tarantool version
	if ctx.Tarantool.TarantoolIsEnterprise {
		tarantoolVersionFileDir := ctx.Tarantool.TarantoolDir
		if ctx.Build.InDocker {
			tarantoolVersionFileDir = ctx.Build.SDKPath
		}

		tarantoolVersionFilePath := filepath.Join(tarantoolVersionFileDir, "VERSION")
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
		tarantoolVersionLine := fmt.Sprintf("TARANTOOL=%s", ctx.Tarantool.TarantoolVersion)
		versionFileLines = append(versionFileLines, tarantoolVersionLine)
	}

	// rocks versions
	rocksVersionsMap, err := common.LuaGetRocksVersions(appDirPath)
	if err != nil {
		log.Warnf("Can't process rocks manifest file. Dependency information can't be "+
			"shipped to the resulting package: %s", err)
	} else {
		for rockName, rockVersion := range rocksVersionsMap {
			if rockName != ctx.Project.Name {
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
