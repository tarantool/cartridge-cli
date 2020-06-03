package project

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/tarantool/cartridge-cli/cli/common"
)

type ProjectCtx struct {
	Name           string
	StateboardName string
	Path           string
	Template       string

	Instances []string
	Daemonize bool

	Verbose bool
	Debug   bool
	Quiet   bool

	PackID  string
	BuildID string

	PackType              string
	TmpDir                string
	PackageFilesDir       string
	BuildDir              string
	ResPackagePath        string
	ResImageTags          []string
	TarantoolDir          string
	TarantoolVersion      string
	TarantoolIsEnterprise bool
	WithStateboard        bool

	BuildInDocker   bool
	BuildFrom       string
	From            string
	DockerNoCache   bool
	DockerCacheFrom []string
	SDKLocal        bool
	SDKPath         string
	BuildSDKDirname string

	Version        string
	Release        string
	VersionRelease string
	Suffix         string
	ImageTags      []string

	UnitTemplatePath          string
	InstUnitTemplatePath      string
	StatboardUnitTemplatePath string

	Entrypoint           string
	StateboardEntrypoint string
	AppDir               string
	ConfPath             string
	RunDir               string
	WorkDirBase          string
}

// FillCtx fills project context
func FillCtx(projectCtx *ProjectCtx) error {
	var err error

	if projectCtx.Path == "" {
		projectCtx.Path, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("Failed to get current directory: %s", err)
		}
	}

	projectCtx.Path, err = filepath.Abs(projectCtx.Path)
	if err != nil {
		return fmt.Errorf("Failed to get absolute path for %s: %s", projectCtx.Path, err)
	}

	if projectCtx.Name == "" {
		if _, err := os.Stat(projectCtx.Path); err != nil {
			return fmt.Errorf("Failed to use specified path: %s", err)
		}
		projectCtx.Name, err = detectName(projectCtx.Path)
		if err != nil {
			return fmt.Errorf(
				"Failed to detect application name: %s. Please pass it explicitly via --name ",
				err,
			)
		}
	}

	projectCtx.StateboardName = fmt.Sprintf("%s-stateboard", projectCtx.Name)

	projectCtx.TarantoolDir, err = common.GetTarantoolDir()
	if err != nil {
		log.Warnf("Failed to find Tarantool executable: %s", err)
	} else {
		projectCtx.TarantoolVersion, err = common.GetTarantoolVersion(projectCtx.TarantoolDir)
		if err != nil {
			return fmt.Errorf("Failed to get Tarantool version: %s", err)
		}

		projectCtx.TarantoolIsEnterprise, err = common.TarantoolIsEnterprise(projectCtx.TarantoolDir)
		if err != nil {
			return fmt.Errorf("Failed to check Tarantool version: %s", err)
		}
	}

	if projectCtx.Entrypoint == "" {
		projectCtx.Entrypoint = defaultEntrypoint
	}

	if projectCtx.StateboardEntrypoint == "" {
		projectCtx.StateboardEntrypoint = defaultStateboardEntrypoint
	}

	projectCtx.AppDir = filepath.Join(defaultAppsDir, projectCtx.Name)

	if projectCtx.ConfPath == "" {
		projectCtx.ConfPath = defaultConfPath
	}

	if projectCtx.RunDir == "" {
		projectCtx.RunDir = defaultRunDir
	}

	if projectCtx.WorkDirBase == "" {
		projectCtx.WorkDirBase = defaultWorkDir
	}

	return nil
}

func detectName(path string) (string, error) {
	var err error

	if _, err := os.Stat(path); err != nil {
		return "", fmt.Errorf("Unable to use specified path: %s", err)
	}

	rockspecPath, err := common.FindRockspec(path)
	if err != nil {
		return "", err
	} else if rockspecPath == "" {
		return "", fmt.Errorf("Application directory should contain rockspec")
	}

	name, err := common.LuaReadStringVar(rockspecPath, "package")
	if err != nil {
		return "", fmt.Errorf("Failed to read `package` field from rockspec: %s", err)
	}

	return name, nil
}
