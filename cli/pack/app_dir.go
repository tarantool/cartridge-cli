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
	"gopkg.in/yaml.v2"

	"github.com/tarantool/cartridge-cli/cli/build"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

const (
	fileReqPerms       = 0444
	dirReqPerms        = 0555
	versionFileName    = "VERSION"
	versionLuaFileName = "VERSION.lua"

	cacheParamsFileName     = "pack-cache.yml"
	maxCachedProjects       = 5
	cntFirstSymbolsFromHash = 10
)

type CachePaths map[string]string

type CachePathParams struct {
	Key         string `yaml:"key,omitempty"`
	KeyPath     string `yaml:"key-path,omitempty"`
	AlwaysCache bool   `yaml:"always-cache,omitempty"`
}

type CachePathsParams map[string]*CachePathParams

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

	log.Debugf("Creating cache directory")
	if err := os.MkdirAll(ctx.Cli.CacheDir, 0755); err != nil {
		return fmt.Errorf("Failed to create cache directory: %s", err)
	}

	cachePaths, err := getProjectCachePaths(ctx)
	if err != nil {
		log.Warnf("%s", err)
	}

	copyFromCache(cachePaths, appDirPath, ctx)

	ctx.Build.Dir = appDirPath
	// Build project
	if err := build.Run(ctx); err != nil {
		return err
	}

	// post-build
	if err := build.PostRun(ctx); err != nil {
		return err
	}

	// Update cache in cartridge temp directory
	if err := updateCache(cachePaths, ctx); err != nil {
		log.Warnf("%s", err)
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

func copyFromCache(paths CachePaths, destPath string, ctx *context.Ctx) {
	if ctx.Pack.NoCache {
		return
	}

	for path, cacheDir := range paths {
		if _, err := os.Stat(cacheDir); err == nil {
			if err := copyPathFromCache(cacheDir, filepath.Join(destPath, path)); err != nil {
				log.Warnf("%s", err)
			}
		} else if !os.IsNotExist(err) {
			log.Warnf("Failed to copy from cache: %s", err)
		}
	}
}

func copyPathFromCache(cachedPath string, destPath string) error {
	basePath := filepath.Base(destPath)
	log.Infof("Using cached path %s", basePath)

	if fileInfo, err := os.Stat(destPath); err == nil {
		if fileInfo.IsDir() {

			if err := copy.Copy(cachedPath, destPath); err != nil {
				log.Warnf("Failed to copy path %s from cache to project directory: %s", destPath, err)
			}
		} else {
			if err := os.MkdirAll(filepath.Base(cachedPath), 0755); err != nil {
				return fmt.Errorf("Failed to create directory: %s", err)
			}

			if err := common.CopyFile(filepath.Join(cachedPath, basePath), destPath); err != nil {
				log.Warnf("Failed to copy path %s from cache to project directory: %s", destPath, err)
			}
		}
	}

	return nil
}

func parseCacheParamsFile(cacheParamsPath string) (*CachePathsParams, error) {
	if _, err := os.Stat(cacheParamsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("File %s with pack cache parameters doesn't exists", cacheParamsPath)
	} else if err != nil {
		return nil, fmt.Errorf("Failed to process %s file: %s", cacheParamsPath, err)
	}

	fileContent, err := common.GetFileContentBytes(cacheParamsPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s file: %s", cacheParamsPath, err)
	}

	var cachePathsParams CachePathsParams
	if err := yaml.Unmarshal(fileContent, &cachePathsParams); err != nil {
		return nil, fmt.Errorf("Failed to parse file %s with pack cache parameters: %s", cacheParamsPath, err)
	}

	return &cachePathsParams, nil
}

func getProjectCachePaths(ctx *context.Ctx) (CachePaths, error) {
	if ctx.Pack.NoCache {
		return nil, nil
	}

	cachePathsParams, err := parseCacheParamsFile(filepath.Join(ctx.Project.Path, cacheParamsFileName))
	if err != nil {
		return nil, err
	}

	cachePaths := CachePaths{}

	for path, params := range *cachePathsParams {
		if params.Key != "" && params.KeyPath != "" {
			// ? or continue
			return nil, fmt.Errorf("You have set both `key` and `key-path` for path %s", path)
		}

		if params.AlwaysCache == true && (params.Key != "" || params.KeyPath != "") {
			// ? or continue
			return nil, fmt.Errorf("You have set to `always-true` flag and have set the hash keys for path %s", path)
		}

		var fullCachePath string
		projectPathHash := common.StringSHA1Hex(ctx.Project.Path)[:cntFirstSymbolsFromHash]

		if params.AlwaysCache == true {
			fullCachePath = filepath.Join(ctx.Cli.CacheDir, projectPathHash, path, "always")
		} else {
			var keyHash string

			if params.KeyPath != "" {
				pathFromProjectRoot := filepath.Join(ctx.Project.Path, params.KeyPath)

				if _, err := os.Stat(pathFromProjectRoot); err == nil {
					if keyHash, err = common.FileSHA1Hex(pathFromProjectRoot); err != nil {
						return nil, fmt.Errorf("Failetd to get hash from file content for path %s: %s", path, err)
					}
				} else if os.IsNotExist(err) {
					return nil, fmt.Errorf("Specified file for the path %s does not exist", path)
				} else {
					return nil, fmt.Errorf("Failed to get specified file for path %s : %s", path, err)
				}
			} else {
				keyHash = common.StringSHA1Hex(params.Key)
			}

			fullCachePath = filepath.Join(ctx.Cli.CacheDir, projectPathHash, path, keyHash[:cntFirstSymbolsFromHash])
		}

		cachePaths[path] = fullCachePath
	}

	return cachePaths, nil
}

func rotateCacheDirs(ctx *context.Ctx) error {
	dir, err := ioutil.ReadDir(ctx.Cli.CacheDir)
	if err != nil {
		return err
	}

	if len(dir) > maxCachedProjects {
		// Sort project cache paths by change time
		sort.Slice(dir, func(i, j int) bool {
			return dir[i].ModTime().Before(dir[j].ModTime())
		})

		for i := 0; i < len(dir)-maxCachedProjects; i++ {
			projectCachePath := filepath.Join(ctx.Cli.CacheDir, dir[i].Name())
			log.Debugf("Removing project cache directory: %s", projectCachePath)
			os.RemoveAll(projectCachePath)
		}
	}

	return nil
}

func updateCache(paths CachePaths, ctx *context.Ctx) error {
	if ctx.Pack.NoCache || paths == nil {
		return nil
	}

	for path, cacheDir := range paths {
		// Delete other caches for this path,
		// because we only store 1 cache for the path
		currentPath := filepath.Dir(cacheDir)

		if _, err := os.Stat(currentPath); err == nil {
			if err := common.ClearDir(currentPath); err != nil {
				log.Warnf("Failed to clear %s cache directory: %s", currentPath, err)
			}
		} else if !os.IsNotExist(err) {
			log.Warnf("Failed to clear %s cache directory: %s", currentPath, err)
		}

		copyPath := filepath.Join(ctx.Build.Dir, path)

		if fileInfo, err := os.Stat(copyPath); err == nil {
			if fileInfo.IsDir() {
				if err := copy.Copy(copyPath, cacheDir); err != nil {
					log.Warnf("Failed to update %s in cache: %s", path, err)
				}
			} else {
				if err := os.MkdirAll(cacheDir, 0755); err != nil {
					return fmt.Errorf("Failed to create cache directory: %s", err)
				}

				if err := common.CopyFile(copyPath, filepath.Join(cacheDir, path)); err != nil {
					log.Warnf("Failed to update %s is cache: %s", path, err)
				}
			}
		}

		log.Debugf("%s cache has been successfully saved in: %s", path, cacheDir)
	}

	return rotateCacheDirs(ctx)
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
