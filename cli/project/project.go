package project

import (
	"fmt"
	"os"
	"path/filepath"

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
	StateboardOnly        bool

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
	AppsDir              string
	AppDir               string
	ConfPath             string
	RunDir               string
	DataDir              string
	LogDir               string
}

func FillTarantoolCtx(projectCtx *ProjectCtx) error {
	var err error

	projectCtx.TarantoolDir, err = common.GetTarantoolDir()
	if err != nil {
		return fmt.Errorf("Failed to find Tarantool executable: %s", err)
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

	return nil
}

func GetStateboardName(projectCtx *ProjectCtx) string {
	return fmt.Sprintf("%s-stateboard", projectCtx.Name)
}

func SetProjectPath(projectCtx *ProjectCtx) error {
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

	return nil
}

func DetectName(path string) (string, error) {
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
