package pack

import (
	"fmt"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/templates"
)

const (
	instanceNameSpecifier = "%i" // https://www.freedesktop.org/software/systemd/man/systemd.unit.html#Specifiers
)

type systemdCtx struct {
	Name           string
	StateboardName string

	DefaultWorkDir    string
	InstanceWorkDir   string
	StateboardWorkDir string

	DefaultPidFile    string
	InstancePidFile   string
	StateboardPidFile string

	DefaultConsoleSock    string
	InstanceConsoleSock   string
	StateboardConsoleSock string

	ConfDir string

	AppEntrypointPath        string
	StateboardEntrypointPath string

	Tarantool string
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

	ctx := getSystemdCtx(projectCtx)

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
			projectCtx.StateboardEntrypoint,
		)
	}

	return &systemdFilesTemplate, nil
}

func getSystemdCtx(projectCtx *project.ProjectCtx) *systemdCtx {
	var ctx systemdCtx

	ctx.Name = projectCtx.Name
	ctx.StateboardName = projectCtx.StateboardName

	ctx.DefaultWorkDir = project.GetInstanceWorkDir(projectCtx, "default")
	ctx.InstanceWorkDir = project.GetInstanceWorkDir(projectCtx, instanceNameSpecifier)
	ctx.StateboardWorkDir = project.GetStateboardWorkDir(projectCtx)

	ctx.DefaultPidFile = project.GetInstancePidFile(projectCtx, "default")
	ctx.InstancePidFile = project.GetInstancePidFile(projectCtx, instanceNameSpecifier)
	ctx.StateboardPidFile = project.GetStateboardPidFile(projectCtx)

	ctx.DefaultConsoleSock = project.GetInstanceConsoleSock(projectCtx, "default")
	ctx.InstanceConsoleSock = project.GetInstanceConsoleSock(projectCtx, instanceNameSpecifier)
	ctx.StateboardConsoleSock = project.GetStateboardConsoleSock(projectCtx)

	ctx.ConfDir = projectCtx.ConfDir

	ctx.AppEntrypointPath = project.GetAppEntrypointPath(projectCtx)
	ctx.StateboardEntrypointPath = project.GetStateboardEntrypointPath(projectCtx)

	if projectCtx.TarantoolIsEnterprise {
		ctx.Tarantool = filepath.Join(projectCtx.AppDir, "tarantool")
	} else {
		ctx.Tarantool = "/usr/bin" // TODO
	}

	return &ctx
}

const (
	appUnitContent = `[Unit]
Description=Tarantool Cartridge app {{ .Name }}.default
After=network.target

[Service]
Type=simple
ExecStartPre=/bin/sh -c 'mkdir -p {{ .DefaultWorkDir }}'
ExecStart={{ .Tarantool }} {{ .AppEntrypointPath }}
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool

Environment=TARANTOOL_APP_NAME={{ .Name }}
Environment=TARANTOOL_WORKDIR={{ .DefaultWorkDir }}
Environment=TARANTOOL_CFG={{ .ConfDir }}
Environment=TARANTOOL_PID_FILE={{ .DefaultPidFile }}
Environment=TARANTOOL_CONSOLE_SOCK={{ .DefaultConsoleSock }}

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
ExecStartPre=/bin/sh -c 'mkdir -p {{ .InstanceWorkDir }}'
ExecStart={{ .Tarantool }} {{ .AppEntrypointPath }}
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool

Environment=TARANTOOL_APP_NAME={{ .Name }}
Environment=TARANTOOL_WORKDIR={{ .InstanceWorkDir }}
Environment=TARANTOOL_CFG={{ .ConfDir }}
Environment=TARANTOOL_PID_FILE={{ .InstancePidFile }}
Environment=TARANTOOL_CONSOLE_SOCK={{ .InstanceConsoleSock }}
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
ExecStart={{ .Tarantool }} {{ .StateboardEntrypointPath }}
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool

Environment=TARANTOOL_APP_NAME={{ .StateboardName }}
Environment=TARANTOOL_WORKDIR={{ .StateboardWorkDir }}
Environment=TARANTOOL_CFG={{ .ConfDir }}
Environment=TARANTOOL_PID_FILE={{ .StateboardPidFile }}
Environment=TARANTOOL_CONSOLE_SOCK={{ .StateboardConsoleSock }}

LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000

# Systemd waits until all xlogs are recovered
TimeoutStartSec=86400s
# Give a reasonable amount of time to close xlogs
TimeoutStopSec=10s

[Install]
WantedBy=multi-user.target
Alias={{ .StateboardName }}
`
)
