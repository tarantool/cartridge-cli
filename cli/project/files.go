package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

const (
	defaultHomeDir = "/home"

	runningConfFilename = ".cartridge.yml"

	defaultEntrypoint           = "init.lua"
	defaultStateboardEntrypoint = "stateboard.init.lua"

	appConfPathSection = "cfg"
	runDirSection      = "run-dir"
	dataDirSection     = "data-dir"
	logDirSection      = "log-dir"
	appsDirSection     = "apps-dir"
	entrypointSection  = "script"
)

var (
	defaultLocalPaths  map[string]string
	defaultGlobalPaths map[string]string
)

func init() {
	defaultLocalPaths = map[string]string{
		appConfPathSection: "instances.yml",
		runDirSection:      "tmp/run",
		dataDirSection:     "tmp/data",
		logDirSection:      "tmp/log",
	}

	defaultGlobalPaths = map[string]string{
		appConfPathSection: "/etc/tarantool/conf.d/",
		runDirSection:      "/var/run/tarantool/",
		dataDirSection:     "/var/lib/tarantool/",
		logDirSection:      "/var/log/tarantool",
		appsDirSection:     "/usr/share/tarantool/",
	}
}

type RunningConf map[string]interface{}

type PathOpts struct {
	SpecifiedPath   string
	ConfSectionName string
	DefaultPath     string
	GetAbs          bool
}

func GetInstanceWorkDir(ctx *context.Ctx, instanceName string) string {
	return filepath.Join(
		ctx.Running.DataDir,
		fmt.Sprintf("%s.%s", ctx.Project.Name, instanceName),
	)
}

func GetStateboardWorkDir(ctx *context.Ctx) string {
	return filepath.Join(
		ctx.Running.DataDir,
		ctx.Project.StateboardName,
	)
}

func GetInstancePidFile(ctx *context.Ctx, instanceName string) string {
	pidFileName := fmt.Sprintf("%s.%s.pid", ctx.Project.Name, instanceName)
	return filepath.Join(
		ctx.Running.RunDir,
		pidFileName,
	)
}

func GetStateboardPidFile(ctx *context.Ctx) string {
	pidFileName := fmt.Sprintf("%s.pid", ctx.Project.StateboardName)
	return filepath.Join(
		ctx.Running.RunDir,
		pidFileName,
	)
}

func GetInstanceConsoleSock(ctx *context.Ctx, instanceName string) string {
	consoleSockName := fmt.Sprintf("%s.%s.control", ctx.Project.Name, instanceName)
	return filepath.Join(
		ctx.Running.RunDir,
		consoleSockName,
	)
}

func GetStateboardConsoleSock(ctx *context.Ctx) string {
	consoleSockName := fmt.Sprintf("%s.control", ctx.Project.StateboardName)
	return filepath.Join(
		ctx.Running.RunDir,
		consoleSockName,
	)
}

func GetInstanceNotifySockPath(ctx *context.Ctx, instanceName string) string {
	notifySockName := fmt.Sprintf("%s.%s.notify", ctx.Project.Name, instanceName)
	return filepath.Join(
		ctx.Running.RunDir,
		notifySockName,
	)
}

func GetStateboardNotifySockPath(ctx *context.Ctx) string {
	notifySockName := fmt.Sprintf("%s.notify", ctx.Project.StateboardName)
	return filepath.Join(
		ctx.Running.RunDir,
		notifySockName,
	)
}

func GetInstanceLogFile(ctx *context.Ctx, instanceName string) string {
	return filepath.Join(
		ctx.Running.LogDir,
		fmt.Sprintf("%s.%s.log", ctx.Project.Name, instanceName),
	)
}

func GetStateboardLogFile(ctx *context.Ctx) string {
	return filepath.Join(
		ctx.Running.LogDir,
		fmt.Sprintf("%s.log", ctx.Project.StateboardName),
	)
}

func GetAppEntrypointPath(ctx *context.Ctx) string {
	return filepath.Join(ctx.Running.AppDir, ctx.Running.Entrypoint)
}

func GetStateboardEntrypointPath(ctx *context.Ctx) string {
	return filepath.Join(ctx.Running.AppDir, ctx.Running.StateboardEntrypoint)
}

func getPath(conf RunningConf, opts PathOpts) (string, error) {
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

	if opts.GetAbs && path != "" {
		if path, err = filepath.Abs(path); err != nil {
			return "", fmt.Errorf("Failed to get absolute path: %s", err)
		}
	}

	return path, nil
}

func setRunningConfPath(ctx *context.Ctx) error {
	if ctx.Running.Global {
		homeDir, err := common.GetHomeDir()
		if err != nil {
			log.Warnf("Failed to get home dir: %s, using the default %s", err, defaultHomeDir)
			homeDir = defaultHomeDir
		}

		ctx.Running.ConfPath = filepath.Join(homeDir, runningConfFilename)
	} else {
		curDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("Failed to get current directory: %s", err)
		}

		ctx.Running.ConfPath = filepath.Join(curDir, runningConfFilename)
	}

	return nil
}

func parseRunningConf(ctx *context.Ctx) (RunningConf, error) {
	if err := setRunningConfPath(ctx); err != nil {
		return nil, fmt.Errorf("Failed to set running conf path: %s", err)
	}

	conf := make(RunningConf)

	if _, err := os.Stat(ctx.Running.ConfPath); err == nil {
		if conf, err = common.ParseYmlFile(ctx.Running.ConfPath); err != nil {
			return nil, fmt.Errorf("Failed to read configuration from file: %s", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("Failed to use conf file: %s", err)
	}

	return conf, nil
}

// SetRunningPaths fills {Run,Data,Log,Conf,Apps,App}Dir
// Values are collected from specified flags and (if useConf is true)
// * ./.cartridge.yml for local running
// * ~/.cartridge.yml for global running
//
// The priority of sources is:
// * user-specified flags
// * value from .cartridge.yml (if useConf is true)
// * default values (defined here in const section)
func SetRunningPaths(ctx *context.Ctx, useConf bool) error {
	var err error
	var conf RunningConf

	if useConf {
		if conf, err = parseRunningConf(ctx); err != nil {
			return fmt.Errorf("Failed to parse conf: %s", err)
		}
	}

	var defaultPaths map[string]string
	if ctx.Running.Global {
		defaultPaths = defaultGlobalPaths
	} else {
		defaultPaths = defaultLocalPaths
	}

	// set directories
	if ctx.Running.Global {
		ctx.Running.AppsDir, err = getPath(nil, PathOpts{
			SpecifiedPath:   ctx.Running.AppsDir,
			ConfSectionName: appsDirSection,
			DefaultPath:     defaultPaths[appsDirSection],
			GetAbs:          true,
		})
		if err != nil {
			return fmt.Errorf("Failed to detect apps dir: %s", err)
		}

		ctx.Running.AppDir = filepath.Join(ctx.Running.AppsDir, ctx.Project.Name)
	}

	ctx.Running.AppConfPath, err = getPath(conf, PathOpts{
		SpecifiedPath:   ctx.Running.AppConfPath,
		ConfSectionName: appConfPathSection,
		DefaultPath:     defaultPaths[appConfPathSection],
		GetAbs:          true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect application conf path: %s", err)
	}

	ctx.Running.RunDir, err = getPath(conf, PathOpts{
		SpecifiedPath:   ctx.Running.RunDir,
		ConfSectionName: runDirSection,
		DefaultPath:     defaultPaths[runDirSection],
		GetAbs:          true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect run dir: %s", err)
	}

	ctx.Running.DataDir, err = getPath(conf, PathOpts{
		SpecifiedPath:   ctx.Running.DataDir,
		ConfSectionName: dataDirSection,
		DefaultPath:     defaultPaths[dataDirSection],
		GetAbs:          true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect data dir: %s", err)
	}

	ctx.Running.LogDir, err = getPath(conf, PathOpts{
		SpecifiedPath:   ctx.Running.LogDir,
		ConfSectionName: logDirSection,
		DefaultPath:     defaultPaths[logDirSection],
		GetAbs:          true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect data dir: %s", err)
	}

	// set entrypoints
	ctx.Running.Entrypoint, err = getPath(conf, PathOpts{
		SpecifiedPath:   ctx.Running.Entrypoint,
		ConfSectionName: entrypointSection,
		DefaultPath:     defaultEntrypoint,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect script: %s", err)
	}

	ctx.Running.StateboardEntrypoint, err = getPath(conf, PathOpts{
		SpecifiedPath: ctx.Running.StateboardEntrypoint,
		DefaultPath:   defaultStateboardEntrypoint,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect stateboard script: %s", err)
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
