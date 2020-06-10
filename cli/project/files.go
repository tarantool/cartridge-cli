package project

import (
	"fmt"
	"path/filepath"
)

const (
	defaultEntrypoint           = "init.lua"
	defaultStateboardEntrypoint = "stateboard.init.lua"

	defaultAppsDir  = "/usr/share/tarantool/"
	defaultConfPath = "/etc/tarantool/conf.d/"
	defaultRunDir   = "/var/run/tarantool/"
	defaultDataDir  = "/var/lib/tarantool/"
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
