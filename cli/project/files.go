package project

import (
	"fmt"
	"path/filepath"
)

const (
	defaultEntrypoint           = "init.lua"
	defaultStateboardEntrypoint = "stateboard.init.lua"

	defaultAppsDir = "/usr/share/tarantool/"
	defaultConfDir = "/etc/tarantool/conf.d/"
	defaultRunDir  = "/var/run/tarantool/"
	defaultWorkDir = "/var/lib/tarantool/"
)

func GetInstanceWorkDir(projectCtx *ProjectCtx, instanceName string) string {
	return filepath.Join(
		projectCtx.WorkDir,
		fmt.Sprintf("%s.%s", projectCtx.Name, instanceName),
	)
}

func GetStateboardWorkDir(projectCtx *ProjectCtx) string {
	return filepath.Join(
		projectCtx.WorkDir,
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

func GetAppEntrypointPath(projectCtx *ProjectCtx) string {
	return filepath.Join(projectCtx.AppDir, projectCtx.Entrypoint)
}

func GetStateboardEntrypointPath(projectCtx *ProjectCtx) string {
	return filepath.Join(projectCtx.AppDir, projectCtx.StateboardEntrypoint)
}
