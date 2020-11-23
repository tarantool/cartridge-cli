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
	cartridgeTmpDirEnv = "CARTRIDGE_TEMPDIR"

	defaultHomeDir             = "/home"
	defaultCartridgeTmpDirName = ".cartridge/tmp"
	tmpSubDirName              = "cartridge.tmp"
	runTmpDirNameFmt           = "run-%s"

	cartridgeLocalConf = ".cartridge.yml"

	defaultEntrypoint           = "init.lua"
	defaultStateboardEntrypoint = "stateboard.init.lua"

	defaultLocalConfPath = "instances.yml"
	defaultLocalDataDir  = "tmp/data"
	defaultLocalLogDir   = "tmp/log"
	defaultLocalAppsDir  = ""

	defaultConfPath = "/etc/tarantool/conf.d/"
	defaultRunDir   = "/var/run/tarantool/"
	defaultDataDir  = "/var/lib/tarantool/"
	defaultLogDir   = "/var/log/tarantool"
	defaultAppsDir  = "/usr/share/tarantool/"

	confPathSection   = "cfg"
	runDirSection     = "run-dir"
	dataDirSection    = "data-dir"
	logDirSection     = "log-dir"
	appsDirSection    = "apps-dir"
	entrypointSection = "script"
)

var (
	defaultCartridgeTmpDir string
)

func init() {
	homeDir, err := common.GetHomeDir()
	if err != nil {
		homeDir = defaultHomeDir
	}

	defaultCartridgeTmpDir = filepath.Join(homeDir, defaultCartridgeTmpDirName)
}

type PathOpts struct {
	SpecifiedPath   string
	ConfSectionName string
	DefaultPath     string
	BasePath        string
}

func GetTmpDirFromEnv(ctx *context.Ctx) {
	ctx.Cli.CartridgeTmpDir = os.Getenv(cartridgeTmpDirEnv)
}

func SetLocalProjectID(ctx *context.Ctx) {
	ctx.Project.ID = common.StringMD5Hex(ctx.Running.AppDir)[:10]
	log.Debugf("Project ID is set to %s", ctx.Project.ID)
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

func getPath(conf map[string]interface{}, opts PathOpts) (string, error) {
	var path string

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

	if path != "" && opts.BasePath != "" && !filepath.IsAbs(path) {
		path = filepath.Join(opts.BasePath, path)
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

	if ctx.Running.AppDir == "" {
		ctx.Running.AppDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("Failed to get current directory: %s", err)
		}
	}

	SetLocalProjectID(ctx)

	conf := make(map[string]interface{})
	cartridgeConfPath := filepath.Join(ctx.Running.AppDir, cartridgeLocalConf)

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
		BasePath:        ctx.Running.AppDir,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect conf path: %s", err)
	}

	log.Debugf("Configuration file path is set to %s", ctx.Running.ConfPath)

	ctx.Running.RunDir, err = getPath(conf, PathOpts{
		SpecifiedPath:   ctx.Running.RunDir,
		ConfSectionName: runDirSection,
		DefaultPath:     "",
		BasePath:        ctx.Running.AppDir,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect run dir: %s", err)
	}

	if ctx.Running.RunDir == "" {
		if err := SetCartridgeTmpDir(ctx); err != nil {
			return fmt.Errorf("Failed to detect tmp directory: %s", err)
		}

		runTmpDirName := fmt.Sprintf(runTmpDirNameFmt, ctx.Project.ID)
		ctx.Running.RunDir = filepath.Join(ctx.Cli.CartridgeTmpDir, runTmpDirName)
	}

	log.Debugf("Run directory is set to %s", ctx.Running.RunDir)

	ctx.Running.DataDir, err = getPath(conf, PathOpts{
		SpecifiedPath:   ctx.Running.DataDir,
		ConfSectionName: dataDirSection,
		DefaultPath:     defaultLocalDataDir,
		BasePath:        ctx.Running.AppDir,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect data dir: %s", err)
	}

	log.Debugf("Data directory is set to %s", ctx.Running.DataDir)

	ctx.Running.LogDir, err = getPath(conf, PathOpts{
		SpecifiedPath:   ctx.Running.LogDir,
		ConfSectionName: logDirSection,
		DefaultPath:     defaultLocalLogDir,
		BasePath:        ctx.Running.AppDir,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect data dir: %s", err)
	}

	log.Debugf("Logs directory is set to %s", ctx.Running.LogDir)

	// set entrypoints
	ctx.Running.Entrypoint, err = getPath(conf, PathOpts{
		SpecifiedPath:   ctx.Running.Entrypoint,
		ConfSectionName: entrypointSection,
		DefaultPath:     defaultEntrypoint,
		BasePath:        ctx.Running.AppDir,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect script: %s", err)
	}

	log.Debugf("Entry point path is set to %s", ctx.Running.Entrypoint)

	ctx.Running.StateboardEntrypoint, err = getPath(conf, PathOpts{
		SpecifiedPath: ctx.Running.StateboardEntrypoint,
		DefaultPath:   defaultStateboardEntrypoint,
		BasePath:      ctx.Running.AppDir,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect stateboard script: %s", err)
	}

	log.Debugf("Stateboard entry point path is set to %s", ctx.Running.StateboardEntrypoint)

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
	})
	if err != nil {
		return fmt.Errorf("Failed to detect apps dir: %s", err)
	}

	ctx.Running.AppDir = filepath.Join(ctx.Running.AppsDir, ctx.Project.Name)

	ctx.Running.ConfPath, err = getPath(nil, PathOpts{
		SpecifiedPath: ctx.Running.ConfPath,
		DefaultPath:   defaultConfPath,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect conf path: %s", err)
	}

	ctx.Running.RunDir, err = getPath(nil, PathOpts{
		SpecifiedPath: ctx.Running.RunDir,
		DefaultPath:   defaultRunDir,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect run dir: %s", err)
	}

	ctx.Running.DataDir, err = getPath(nil, PathOpts{
		SpecifiedPath: ctx.Running.DataDir,
		DefaultPath:   defaultDataDir,
	})
	if err != nil {
		return fmt.Errorf("Failed to detect data dir: %s", err)
	}

	ctx.Running.LogDir, err = getPath(nil, PathOpts{
		SpecifiedPath: ctx.Running.LogDir,
		DefaultPath:   defaultLogDir,
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

func SetCartridgeTmpDir(ctx *context.Ctx) error {
	var err error

	if ctx.Cli.CartridgeTmpDir == "" {
		// tmp dir wasn't specified
		ctx.Cli.CartridgeTmpDir = defaultCartridgeTmpDir
	} else {
		// tmp dir was specified
		ctx.Cli.CartridgeTmpDir, err = filepath.Abs(ctx.Cli.CartridgeTmpDir)
		if err != nil {
			return fmt.Errorf(
				"Failed to get absolute path for specified temporary dir %s: %s",
				ctx.Cli.CartridgeTmpDir,
				err,
			)
		}

		if fileInfo, err := os.Stat(ctx.Cli.CartridgeTmpDir); err == nil {
			// directory is already exists

			if !fileInfo.IsDir() {
				return fmt.Errorf(
					"Specified temporary directory is not a directory: %s",
					ctx.Cli.CartridgeTmpDir,
				)
			}

			// This little hack is used to prevent deletion of user files
			// from the specified tmp directory on cleanup.
			ctx.Cli.CartridgeTmpDir = filepath.Join(ctx.Cli.CartridgeTmpDir, tmpSubDirName)

		} else if !os.IsNotExist(err) {
			return fmt.Errorf(
				"Unable to use specified temporary directory %s: %s",
				ctx.Cli.CartridgeTmpDir,
				err,
			)
		}
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
