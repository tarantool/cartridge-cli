package project

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
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
)

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

func SetLocalRunningPaths(projectCtx *ProjectCtx) error {
	curDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Failed to get current directory: %s", err)
	}

	if projectCtx.RunDir == "" {
		projectCtx.RunDir = filepath.Join(curDir, defaultLocalRunDir)
	}
	if projectCtx.RunDir, err = filepath.Abs(projectCtx.RunDir); err != nil {
		return fmt.Errorf("Failed to get run dir absolute path: %s", err)
	}

	if projectCtx.DataDir == "" {
		projectCtx.DataDir = filepath.Join(curDir, defaultLocalDataDir)
	}
	if projectCtx.DataDir, err = filepath.Abs(projectCtx.DataDir); err != nil {
		return fmt.Errorf("Failed to get data dir absolute path: %s", err)
	}

	if projectCtx.LogDir == "" {
		projectCtx.LogDir = filepath.Join(curDir, defaultLocalLogDir)
	}
	if projectCtx.LogDir, err = filepath.Abs(projectCtx.LogDir); err != nil {
		return fmt.Errorf("Failed to get log dir absolute path: %s", err)
	}

	if projectCtx.ConfPath == "" {
		projectCtx.ConfPath = filepath.Join(curDir, defaultLocalConfPath)
	}
	if projectCtx.ConfPath, err = filepath.Abs(projectCtx.ConfPath); err != nil {
		return fmt.Errorf("Failed to get conf path absolute path: %s", err)
	}

	return nil
}

func SetSystemRunningPaths(projectCtx *ProjectCtx) error {
	projectCtx.AppDir = filepath.Join(defaultAppsDir, projectCtx.Name)

	if projectCtx.ConfPath == "" {
		projectCtx.ConfPath = defaultConfPath
	}

	if projectCtx.RunDir == "" {
		projectCtx.RunDir = defaultRunDir
	}

	if projectCtx.DataDir == "" {
		projectCtx.DataDir = defaultDataDir
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
