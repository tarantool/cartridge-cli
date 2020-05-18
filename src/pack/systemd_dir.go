package pack

import (
	"fmt"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/tarantool/cartridge-cli/project"
	"github.com/tarantool/cartridge-cli/templates"
)

type systemdCtx struct {
	AppName       string
	AppDir        string
	AppWorkDir    string
	AppEntrypoint string

	TarantoolDir string

	StateboardName       string
	StateboardWorkDir    string
	StateboardEntrypoint string
}

var (
	systemdAppFilesTemplate = templates.FileTreeTemplate{
		Dirs: []templates.DirTemplate{
			{
				Path: "/etc/systemd/system/",
				Mode: 0755,
			},
		},
		Files: []templates.FileTemplate{
			{
				Path:    "/etc/systemd/system/{{ .AppName }}.service",
				Mode:    0644,
				Content: appUnitContent,
			},
			{
				Path:    "/etc/systemd/system/{{ .AppName }}@.service",
				Mode:    0644,
				Content: appInstUnitContent,
			},
		},
	}

	systemdStateboardFilesTemplate = templates.FileTreeTemplate{
		Dirs: []templates.DirTemplate{
			{
				Path: "/etc/systemd/system/",
				Mode: 0755,
			},
		},
		Files: []templates.FileTemplate{
			{
				Path:    "/etc/systemd/system/{{ .StateboardName }}.service",
				Mode:    0644,
				Content: stateboardUnitContent,
			},
		},
	}
)

func initSystemdDir(baseDirPath string, projectCtx *project.ProjectCtx) error {
	log.Debugf("Create systemd dir in %s", baseDirPath)

	systemdFilesTemplate := systemdAppFilesTemplate
	if projectCtx.WithStateboard {
		systemdFilesTemplate = *templates.Combine(
			systemdFilesTemplate,
			systemdStateboardFilesTemplate,
		)
	}

	ctx := systemdCtx{
		AppName:       projectCtx.Name,
		AppDir:        filepath.Join("/usr/share/tarantool", projectCtx.Name),
		AppWorkDir:    filepath.Join("/var/lib/tarantool/", projectCtx.Name),
		AppEntrypoint: project.AppEntrypointName,

		StateboardName:       projectCtx.StateboardName,
		StateboardWorkDir:    filepath.Join("/var/lib/tarantool", projectCtx.StateboardName),
		StateboardEntrypoint: project.StateboardEntrypointName,
	}

	if projectCtx.TarantoolIsEnterprise {
		ctx.TarantoolDir = ctx.AppDir
	} else {
		ctx.TarantoolDir = "usr/bin" // TODO
	}

	// TODO: use custom unit files

	if err := templates.InstantiateTree(&systemdFilesTemplate, baseDirPath, ctx); err != nil {
		return fmt.Errorf("Failed to instantiate systemd dir: %s", err)
	}

	return nil
}

const (
	appUnitContent = `[Unit]
Description=Tarantool Cartridge app {{ .AppName }}.default
After=network.target
[Service]
Type=simple
ExecStartPre=/bin/sh -c 'mkdir -p {{ .AppWorkDir }}.default'
ExecStart={{ .TarantoolDir }}/tarantool {{ .AppDir }}/{{ .AppEntrypoint }}
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool
Environment=TARANTOOL_APP_NAME={{ .AppName }}
Environment=TARANTOOL_WORKDIR={{ .AppWorkDir }}.default
Environment=TARANTOOL_CFG=/etc/tarantool/conf.d/
Environment=TARANTOOL_PID_FILE=/var/run/tarantool/{{ .AppName }}.default.pid
Environment=TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/{{ .AppName }}.default.control
LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Increase fd limit for Vinyl
LimitNOFILE=65535
# Systemd waits until all xlogs are recovered
TimeoutStartSec=86400s
# Give a reasonable amount of time to close xlogs
TimeoutStopSec=10s
[Install]
WantedBy=multi-user.target
Alias={{ .AppName }}
`
	appInstUnitContent = `[Unit]
Description=Tarantool Cartridge app {{ .AppName }}@%i
After=network.target
[Service]
Type=simple
ExecStartPre=/bin/sh -c 'mkdir -p {{ .AppWorkDir }}.%i'
ExecStart={{ .TarantoolDir }}/tarantool {{ .AppDir }}/{{ .AppEntrypoint }}
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool
Environment=TARANTOOL_APP_NAME={{ .AppName }}
Environment=TARANTOOL_WORKDIR={{ .AppWorkDir }}.%i
Environment=TARANTOOL_CFG=/etc/tarantool/conf.d/
Environment=TARANTOOL_PID_FILE=/var/run/tarantool/{{ .AppName }}.%i.pid
Environment=TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/{{ .AppName }}.%i.control
Environment=TARANTOOL_INSTANCE_NAME=%i
LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Increase fd limit for Vinyl
LimitNOFILE=65535
# Systemd waits until all xlogs are recovered
TimeoutStartSec=86400s
# Give a reasonable amount of time to close xlogs
TimeoutStopSec=10s
[Install]
WantedBy=multi-user.target
Alias={{ .AppName }}.%i
`
	stateboardUnitContent = `[Unit]
Description=Tarantool Cartridge stateboard for {{ .AppName }}
After=network.target
[Service]
Type=simple
ExecStartPre=/bin/sh -c 'mkdir -p {{ .StateboardWorkDir }}'
ExecStart={{ .TarantoolDir }}/tarantool {{ .AppDir }}/{{ .StateboardEntrypoint }}
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool
Environment=TARANTOOL_APP_NAME={{ .StateboardName }}
Environment=TARANTOOL_WORKDIR={{ .StateboardWorkDir }}
Environment=TARANTOOL_CFG=/etc/tarantool/conf.d/
Environment=TARANTOOL_PID_FILE=/var/run/tarantool/{{ .StateboardName }}.pid
Environment=TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/{{ .StateboardName }}.control
LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Systemd waits until all xlogs are recovered
TimeoutStartSec=86400s
# Give a reasonable amount of time to close xlogs
TimeoutStopSec=10s
[Install]
WantedBy=multi-user.target
Alias={{ .StateboardName}}
`
)
