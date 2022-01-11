package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

const (
	cartridgeLocalConf = ".cartridge.yml"

	defaultEntrypoint           = "init.lua"
	defaultStateboardEntrypoint = "stateboard.init.lua"

	defaultLocalConfPath = "instances.yml"
	defaultLocalRunDir   = "tmp/run"
	defaultLocalDataDir  = "tmp/data"
	defaultLocalLogDir   = "tmp/log"
	defaultLocalAppsDir  = ""

	defaultConfPath       = "/etc/tarantool/conf.d/"
	defaultRunDir         = "/var/run/tarantool/"
	defaultDataDir        = "/var/lib/tarantool/"
	defaultLogDir         = "/var/log/tarantool"
	defaultAppsDir        = "/usr/share/tarantool/"
	defaultStateboardFlag = false

	confPathSection       = "cfg"
	runDirSection         = "run-dir"
	dataDirSection        = "data-dir"
	logDirSection         = "log-dir"
	appsDirSection        = "apps-dir"
	entrypointSection     = "script"
	confStateboardSection = "stateboard"
)

type PathOpts struct {
	SpecifiedPath   string
	ConfSectionName string
	DefaultPath     string
	GetAbs          bool
}

type FlagOpts struct {
	SpecifiedFlag   bool
	ConfSectionName string
	DefaultFlag     bool
	FlagIsSet       bool
}

func GetInstanceID(ctx *context.Ctx, instanceName string) string {
	return fmt.Sprintf("%s.%s", ctx.Project.Name, instanceName)
}

func GetInstanceWorkDir(ctx *context.Ctx, instanceName string) string {
	return filepath.Join(
		ctx.Running.DataDir,
		GetInstanceID(ctx, instanceName),
	)
}

func GetStateboardWorkDir(ctx *context.Ctx) string {
	return filepath.Join(
		ctx.Running.DataDir,
		ctx.Project.StateboardName,
	)
}

func GetInstancePidFile(ctx *context.Ctx, instanceName string) string {
	pidFileName := fmt.Sprintf("%s.pid", GetInstanceID(ctx, instanceName))
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
	consoleSockName := fmt.Sprintf("%s.control", GetInstanceID(ctx, instanceName))
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
	notifySockName := fmt.Sprintf("%s.notify", GetInstanceID(ctx, instanceName))
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
		fmt.Sprintf("%s.log", GetInstanceID(ctx, instanceName)),
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

func getFlag(conf map[string]interface{}, opts FlagOpts) (bool, error) {
	var flag bool

	if opts.FlagIsSet {
		flag = opts.SpecifiedFlag
	} else if value, found := conf[opts.ConfSectionName]; found {
		var ok bool
		if flag, ok = value.(bool); !ok {
			return false, fmt.Errorf("%s value should be `true` or `false`", opts.ConfSectionName)
		}
	} else {
		flag = opts.DefaultFlag
	}

	return flag, nil
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

	if opts.GetAbs && path != "" {
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
func SetLocalRunningPaths(ctx *context.Ctx) error {
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
	ctx.Running.ConfPath, err = getPath(conf, PathOpts{
		SpecifiedPath:   ctx.Running.ConfPath,
		ConfSectionName: confPathSection,
		DefaultPath:     defaultLocalConfPath,
		GetAbs:          true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect conf path: %s", err)
	}

	ctx.Running.RunDir, err = getPath(conf, PathOpts{
		SpecifiedPath:   ctx.Running.RunDir,
		ConfSectionName: runDirSection,
		DefaultPath:     defaultLocalRunDir,
		GetAbs:          true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect run dir: %s", err)
	}

	ctx.Running.DataDir, err = getPath(conf, PathOpts{
		SpecifiedPath:   ctx.Running.DataDir,
		ConfSectionName: dataDirSection,
		DefaultPath:     defaultLocalDataDir,
		GetAbs:          true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect data dir: %s", err)
	}

	ctx.Running.LogDir, err = getPath(conf, PathOpts{
		SpecifiedPath:   ctx.Running.LogDir,
		ConfSectionName: logDirSection,
		DefaultPath:     defaultLocalLogDir,
		GetAbs:          true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect log dir: %s", err)
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

	// set stateboard flag
	ctx.Running.WithStateboard, err = getFlag(conf, FlagOpts{
		SpecifiedFlag:   ctx.Running.WithStateboard,
		ConfSectionName: confStateboardSection,
		DefaultFlag:     defaultStateboardFlag,
		FlagIsSet:       ctx.Running.StateboardFlagIsSet,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect stateboard flag: %s", err)
	}

	return nil
}

// SetSystemRunningPaths fills {Run,Data,Log,Conf}Dir
// Values are collected from specified flags
//
// The priority of sources is:
// * user-specified flags
// * default values (defined here in const section)
func SetSystemRunningPaths(ctx *context.Ctx) error {
	var err error

	// set directories
	ctx.Running.AppsDir, err = getPath(nil, PathOpts{
		SpecifiedPath: ctx.Running.AppsDir,
		DefaultPath:   defaultAppsDir,
		GetAbs:        true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect apps dir: %s", err)
	}

	ctx.Running.AppDir = filepath.Join(ctx.Running.AppsDir, ctx.Project.Name)

	ctx.Running.ConfPath, err = getPath(nil, PathOpts{
		SpecifiedPath: ctx.Running.ConfPath,
		DefaultPath:   defaultConfPath,
		GetAbs:        true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect conf path: %s", err)
	}

	ctx.Running.RunDir, err = getPath(nil, PathOpts{
		SpecifiedPath: ctx.Running.RunDir,
		DefaultPath:   defaultRunDir,
		GetAbs:        true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect run dir: %s", err)
	}

	ctx.Running.DataDir, err = getPath(nil, PathOpts{
		SpecifiedPath: ctx.Running.DataDir,
		DefaultPath:   defaultDataDir,
		GetAbs:        true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect data dir: %s", err)
	}

	ctx.Running.LogDir, err = getPath(nil, PathOpts{
		SpecifiedPath: ctx.Running.LogDir,
		DefaultPath:   defaultLogDir,
		GetAbs:        true,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect log dir: %s", err)
	}

	// set entrypoints
	ctx.Running.Entrypoint, err = getPath(nil, PathOpts{
		SpecifiedPath: ctx.Running.Entrypoint,
		DefaultPath:   defaultEntrypoint,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect entrypoint: %s", err)
	}

	ctx.Running.StateboardEntrypoint, err = getPath(nil, PathOpts{
		SpecifiedPath: ctx.Running.StateboardEntrypoint,
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
