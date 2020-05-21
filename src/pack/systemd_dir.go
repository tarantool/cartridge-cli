package pack

import (
	"fmt"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/tarantool/cartridge-cli/src/common"
	"github.com/tarantool/cartridge-cli/src/project"
	"github.com/tarantool/cartridge-cli/src/templates"
)

type systemdCtx struct {
	Name       string
	AppDir     string
	WorkDir    string
	Entrypoint string

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
		Files: []templates.FileTemplate{},
	}

	defaultAppUnitTemplate = templates.FileTemplate{
		Path:    "/etc/systemd/system/{{ .Name }}.service",
		Mode:    0644,
		Content: appUnitContent,
	}

	defaultAppInstUnitTemplate = templates.FileTemplate{
		Path:    "/etc/systemd/system/{{ .Name }}@.service",
		Mode:    0644,
		Content: appInstUnitContent,
	}

	defaultStateboardUnitTemplate = templates.FileTemplate{
		Path:    "/etc/systemd/system/{{ .StateboardName }}.service",
		Mode:    0644,
		Content: stateboardUnitContent,
	}
)

func initSystemdDir(baseDirPath string, projectCtx *project.ProjectCtx) error {
	log.Debugf("Create systemd dir in %s", baseDirPath)

	ctx := systemdCtx{
		Name:       projectCtx.Name,
		AppDir:     filepath.Join("/usr/share/tarantool", projectCtx.Name),
		WorkDir:    filepath.Join("/var/lib/tarantool/", projectCtx.Name),
		Entrypoint: project.AppEntrypointName,

		StateboardName:       projectCtx.StateboardName,
		StateboardWorkDir:    filepath.Join("/var/lib/tarantool", projectCtx.StateboardName),
		StateboardEntrypoint: project.StateboardEntrypointName,
	}

	if projectCtx.TarantoolIsEnterprise {
		ctx.TarantoolDir = ctx.AppDir
	} else {
		ctx.TarantoolDir = "/usr/bin" // TODO
	}

	systemdFilesTemplate, err := getSystemdTemplate(projectCtx)
	if err != nil {
		return err
	}

	if err := systemdFilesTemplate.Instantiate(baseDirPath, ctx); err != nil {
		return fmt.Errorf("Failed to instantiate systemd dir: %s", err)
	}

	return nil
}

func getSystemdTemplate(projectCtx *project.ProjectCtx) (templates.Template, error) {
	var err error

	systemdFilesTemplate := systemdAppFilesTemplate

	// app unit file template
	appUnit := defaultAppUnitTemplate
	if projectCtx.UnitTemplatePath != "" {
		appUnit.Content, err = common.GetFileContent(projectCtx.UnitTemplatePath)
		if err != nil {
			return nil, fmt.Errorf("Failed to read specified unit template: %s", err)
		}
	}

	systemdFilesTemplate.AddFiles(appUnit)

	// app instantiated unit file template
	appInstUnit := defaultAppInstUnitTemplate
	if projectCtx.InstUnitTemplatePath != "" {
		appInstUnit.Content, err = common.GetFileContent(projectCtx.InstUnitTemplatePath)
		if err != nil {
			return nil, fmt.Errorf("Failed to read specified instantiated unit template: %s", err)
		}
	}

	systemdFilesTemplate.AddFiles(appInstUnit)

	// stateboard unit file template
	if projectCtx.WithStateboard {
		stateboardUnit := defaultStateboardUnitTemplate
		if projectCtx.StatboardUnitTemplatePath != "" {
			stateboardUnit.Content, err = common.GetFileContent(projectCtx.StatboardUnitTemplatePath)
			if err != nil {
				return nil, fmt.Errorf("Failed to read specified stateboard unit template: %s", err)
			}
		}
		systemdFilesTemplate.AddFiles(stateboardUnit)
	} else {
		log.Warnf(
			"App directory doesn't contain stateboard entrypoint script `%s`. "+
				"Stateboard systemd service unit file wouldn't be delivered",
			project.StateboardEntrypointName,
		)
	}

	return &systemdFilesTemplate, nil
}

const (
	appUnitContent = `[Unit]
Description=Tarantool Cartridge app {{ .Name }}.default
After=network.target
[Service]
Type=simple
ExecStartPre=/bin/sh -c 'mkdir -p {{ .WorkDir }}.default'
ExecStart={{ .TarantoolDir }}/tarantool {{ .AppDir }}/{{ .Entrypoint }}
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool
Environment=TARANTOOL_APP_NAME={{ .Name }}
Environment=TARANTOOL_WORKDIR={{ .WorkDir }}.default
Environment=TARANTOOL_CFG=/etc/tarantool/conf.d/
Environment=TARANTOOL_PID_FILE=/var/run/tarantool/{{ .Name }}.default.pid
Environment=TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/{{ .Name }}.default.control
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
Alias={{ .Name }}
`
	appInstUnitContent = `[Unit]
Description=Tarantool Cartridge app {{ .Name }}@%i
After=network.target
[Service]
Type=simple
ExecStartPre=/bin/sh -c 'mkdir -p {{ .WorkDir }}.%i'
ExecStart={{ .TarantoolDir }}/tarantool {{ .AppDir }}/{{ .Entrypoint }}
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool
Environment=TARANTOOL_APP_NAME={{ .Name }}
Environment=TARANTOOL_WORKDIR={{ .WorkDir }}.%i
Environment=TARANTOOL_CFG=/etc/tarantool/conf.d/
Environment=TARANTOOL_PID_FILE=/var/run/tarantool/{{ .Name }}.%i.pid
Environment=TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/{{ .Name }}.%i.control
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
Alias={{ .Name }}.%i
`
	stateboardUnitContent = `[Unit]
Description=Tarantool Cartridge stateboard for {{ .Name }}
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
