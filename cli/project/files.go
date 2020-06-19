package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tarantool/cartridge-cli/cli/common"
)

const (
	cartridgeLocalConf = ".cartridge.yml"

	defaultEntrypoint           = "init.lua"
	defaultStateboardEntrypoint = "stateboard.init.lua"

	defaultLocalConfPath = "instances.yml"
	defaultLocalRunDir   = "tmp/run"
	defaultLocalDataDir  = "tmp/data"
	defaultLocalLogDir   = "tmp/log"

	defaultConfPath = "/etc/tarantool/conf.d/"
	defaultRunDir   = "/var/run/tarantool/"
	defaultDataDir  = "/var/lib/tarantool/"
	defaultAppsDir  = "/usr/share/tarantool/"

	confPathSection   = "cfg"
	runDirSection     = "run-dir"
	dataDirSection    = "data-dir"
	logDirSection     = "log-dir"
	entrypointSection = "script"
)

type PathOpts struct {
	SpecifiedPath   string
	ConfSectionName string
	DefaultPath     string
	GetAbs          bool
}

func GetInstanceWorkDir(projectCtx *ProjectCtx, instanceName string) string {
	return filepath.Join(
		projectCtx.DataDir,
		fmt.Sprintf("%s.%s", projectCtx.Name, instanceName),
	)
}

func GetStateboardWorkDir(projectCtx *ProjectCtx) string {
	return filepath.Join(
		projectCtx.DataDir,
		projectCtx.StateboardName,
	)
}

func GetInstancePidFile(projectCtx *ProjectCtx, instanceName string) string {
	pidFileName := fmt.Sprintf("%s.%s.pid", projectCtx.Name, instanceName)
	return filepath.Join(
		projectCtx.RunDir,
		pidFileName,
	)
}

func GetStateboardPidFile(projectCtx *ProjectCtx) string {
	pidFileName := fmt.Sprintf("%s.pid", projectCtx.StateboardName)
	return filepath.Join(
		projectCtx.RunDir,
		pidFileName,
	)
}

func GetInstanceConsoleSock(projectCtx *ProjectCtx, instanceName string) string {
	consoleSockName := fmt.Sprintf("%s.%s.control", projectCtx.Name, instanceName)
	return filepath.Join(
		projectCtx.RunDir,
		consoleSockName,
	)
}

func GetStateboardConsoleSock(projectCtx *ProjectCtx) string {
	consoleSockName := fmt.Sprintf("%s.control", projectCtx.StateboardName)
	return filepath.Join(
		projectCtx.RunDir,
		consoleSockName,
	)
}

func GetInstanceNotifySockPath(projectCtx *ProjectCtx, instanceName string) string {
	notifySockName := fmt.Sprintf("%s.%s.notify", projectCtx.Name, instanceName)
	return filepath.Join(
		projectCtx.RunDir,
		notifySockName,
	)
}

func GetStateboardNotifySockPath(projectCtx *ProjectCtx) string {
	notifySockName := fmt.Sprintf("%s.notify", projectCtx.StateboardName)
	return filepath.Join(
		projectCtx.RunDir,
		notifySockName,
	)
}

func GetInstanceLogFile(projectCtx *ProjectCtx, instanceName string) string {
	return filepath.Join(
		projectCtx.LogDir,
		fmt.Sprintf("%s.%s.log", projectCtx.Name, instanceName),
	)
}

func GetStateboardLogFile(projectCtx *ProjectCtx) string {
	return filepath.Join(
		projectCtx.LogDir,
		fmt.Sprintf("%s.log", projectCtx.StateboardName),
	)
}

func GetAppEntrypointPath(projectCtx *ProjectCtx) string {
	return filepath.Join(projectCtx.AppDir, projectCtx.Entrypoint)
}

func GetStateboardEntrypointPath(projectCtx *ProjectCtx) string {
	return filepath.Join(projectCtx.AppDir, projectCtx.StateboardEntrypoint)
}

func getPath(conf map[string]interface{}, opts PathOpts) (string, error) {
	var path string
	var err error

	if opts.SpecifiedPath != "" {
		path = opts.SpecifiedPath
	} else if conf == nil || opts.ConfSectionName == "" {
		path = opts.DefaultPath
	} else if pathFromConf, found := conf[opts.ConfSectionName]; found {
		var ok bool
		if path, ok = pathFromConf.(string); !ok {
			return "", fmt.Errorf("%s config value should be string", opts.ConfSectionName)
		}
	} else {
		path = opts.DefaultPath
	}

	if opts.GetAbs {
		if path, err = filepath.Abs(path); err != nil {
			return "", fmt.Errorf("Failed to get absolute path: %s", err)
		}
	}

	return path, nil
}

// SetLocalRunningPaths fills {Run,Data,Log,Conf}Dir
// Values are collected from specified flags and .cartridge.yml
//
// The priority of sources is:
// * user-specified flags
// * value from .cartridge.yml
// * default values (defined here in const section)
func SetLocalRunningPaths(projectCtx *ProjectCtx) error {
	var err error

	curDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Failed to get current directory: %s", err)
	}

	conf := make(map[string]interface{})
	cartridgeConfPath := filepath.Join(curDir, cartridgeLocalConf)

	if _, err := os.Stat(cartridgeConfPath); err == nil {
		if conf, err = common.ParseYmlFile(cartridgeConfPath); err != nil {
			return fmt.Errorf("Failed to read configuration from file: %s", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Failed to use conf file: %s", err)
	}

	// set directories
	projectCtx.ConfPath, err = getPath(conf, PathOpts{
		SpecifiedPath:   projectCtx.ConfPath,
		ConfSectionName: confPathSection,
		DefaultPath:     defaultLocalConfPath,
		GetAbs:          true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect conf path: %s", err)
	}

	projectCtx.RunDir, err = getPath(conf, PathOpts{
		SpecifiedPath:   projectCtx.RunDir,
		ConfSectionName: runDirSection,
		DefaultPath:     defaultLocalRunDir,
		GetAbs:          true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect run dir: %s", err)
	}

	projectCtx.DataDir, err = getPath(conf, PathOpts{
		SpecifiedPath:   projectCtx.DataDir,
		ConfSectionName: dataDirSection,
		DefaultPath:     defaultLocalDataDir,
		GetAbs:          true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect data dir: %s", err)
	}

	projectCtx.LogDir, err = getPath(conf, PathOpts{
		SpecifiedPath:   projectCtx.LogDir,
		ConfSectionName: logDirSection,
		DefaultPath:     defaultLocalLogDir,
		GetAbs:          true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect data dir: %s", err)
	}

	// set entrypoints
	projectCtx.Entrypoint, err = getPath(conf, PathOpts{
		SpecifiedPath:   projectCtx.Entrypoint,
		ConfSectionName: entrypointSection,
		DefaultPath:     defaultEntrypoint,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect script: %s", err)
	}

	projectCtx.StateboardEntrypoint, err = getPath(conf, PathOpts{
		SpecifiedPath: projectCtx.StateboardEntrypoint,
		DefaultPath:   defaultStateboardEntrypoint,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect stateboard script: %s", err)
	}

	return nil
}

func SetSystemRunningPaths(projectCtx *ProjectCtx) error {
	var err error

	projectCtx.AppDir = filepath.Join(defaultAppsDir, projectCtx.Name)

	// set directories
	projectCtx.ConfPath, err = getPath(nil, PathOpts{
		SpecifiedPath: projectCtx.ConfPath,
		DefaultPath:   defaultConfPath,
		GetAbs:        true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect conf path: %s", err)
	}

	projectCtx.RunDir, err = getPath(nil, PathOpts{
		SpecifiedPath: projectCtx.RunDir,
		DefaultPath:   defaultRunDir,
		GetAbs:        true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect run dir: %s", err)
	}

	projectCtx.DataDir, err = getPath(nil, PathOpts{
		SpecifiedPath: projectCtx.DataDir,
		DefaultPath:   defaultDataDir,
		GetAbs:        true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect data dir: %s", err)
	}

	// set entrypoints
	projectCtx.Entrypoint, err = getPath(nil, PathOpts{
		SpecifiedPath: projectCtx.Entrypoint,
		DefaultPath:   defaultEntrypoint,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect entrypoint: %s", err)
	}

	projectCtx.StateboardEntrypoint, err = getPath(nil, PathOpts{
		SpecifiedPath: projectCtx.StateboardEntrypoint,
		DefaultPath:   defaultStateboardEntrypoint,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect stateboard entrypoint: %s", err)
	}

	return nil
}

const (
	PreInstScriptContent = `/bin/sh -c 'groupadd -r tarantool > /dev/null 2>&1 || :'
/bin/sh -c 'useradd -M -N -g tarantool -r -d /var/lib/tarantool -s /sbin/nologin \
    -c "Tarantool Server" tarantool > /dev/null 2>&1 || :'
/bin/sh -c 'mkdir -p /etc/tarantool/conf.d/ --mode 755 2>&1 || :'
/bin/sh -c 'mkdir -p /var/lib/tarantool/ --mode 755 2>&1 || :'
/bin/sh -c 'chown tarantool:tarantool /var/lib/tarantool 2>&1 || :'
/bin/sh -c 'mkdir -p /var/run/tarantool/ --mode 755 2>&1 || :'
/bin/sh -c 'chown tarantool:tarantool /var/run/tarantool 2>&1 || :'
`

	PostInstScriptContent = `
/bin/sh -c 'chown -R root:root /usr/share/tarantool/{{ .Name }}'
/bin/sh -c 'chown root:root /etc/systemd/system/{{ .Name }}.service'
/bin/sh -c 'chown root:root /etc/systemd/system/{{ .Name }}@.service'
/bin/sh -c 'chown root:root /usr/lib/tmpfiles.d/{{ .Name }}.conf'
`
)
