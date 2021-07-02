package pack

import (
	"fmt"
	"os"
	"path/filepath"
	"gopkg.in/yaml.v2"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/templates"
)

const (
	instanceNameSpecifier = "%i" // https://www.freedesktop.org/software/systemd/man/systemd.unit.html#Specifiers

	defaultSystemdUnitParamsFileName = "systemd-unit-params.yml"

	minFdLimit = 1024
	minStateboardFdLimit = 1024

	defaultInstanceFdLimit = 65535
	defaultStateboardFdLimit = 65535
)

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

type SystemdUnitParams struct {
	FdLimit           *int `yaml:"fd-limit"`
	StateboardFdLimit *int `yaml:"stateboard-fd-limit"`
}

func parseSystemdUnitParamsFile(systemdUnitParamsPath string, defaultUnitParamsPath string) (*SystemdUnitParams, error) {
	var fileContentBytes []byte
	var err error

	if systemdUnitParamsPath == "" {
		if _, err := os.Stat(defaultUnitParamsPath); err == nil {
			log.Debugf("Default file with system unit params is used: %s", systemdUnitParamsPath)
			systemdUnitParamsPath = defaultUnitParamsPath
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("Failed to use default file with system unit params: %s", err)
		}
	}

	if systemdUnitParamsPath == "" {
		return &SystemdUnitParams{}, nil
	}

	if _, err := os.Stat(systemdUnitParamsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Specified file with system unit params %s doesn't exists", systemdUnitParamsPath)
	} else if err != nil {
		return nil, fmt.Errorf("Impossible to use specified file %s: %s", systemdUnitParamsPath, err)
	}

	fileContentBytes, err = common.GetFileContentBytes(systemdUnitParamsPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file with system unit params:  %s", err)
	}

	var systemdUnitParams SystemdUnitParams
	if err := yaml.Unmarshal([]byte(fileContentBytes), &systemdUnitParams); err != nil {
		return nil, fmt.Errorf("Failed to parse system unit params file %s: %s", systemdUnitParamsPath, err)
	}

	return &systemdUnitParams, nil
}

func checkMinValue(paramName string, value int, minValue int) error {
	if value < minValue {
		return fmt.Errorf("Incorrect value for %s: minimal value is %d", paramName, minValue)
	}

	return nil
}

func setDefaults(valuePtr **int, defaultValue int) error {
	if valuePtr == nil {
		return project.InternalError("Failed to set default value: memory allocation error")
	}

	if *valuePtr == nil {
		*valuePtr = new(int)
		**valuePtr = defaultValue
	}

	return nil
}

func getSystemdUnitParams(ctx *context.Ctx) (*SystemdUnitParams, error) {
	var err error

	systemdUnitParams, err := parseSystemdUnitParamsFile(
		ctx.Pack.SystemdUnitParamsPath,
		filepath.Join(ctx.Project.Path, defaultSystemdUnitParamsFileName),
	)
	if err != nil {
		return nil, err
	}

	if err := setDefaults(&systemdUnitParams.FdLimit, defaultInstanceFdLimit); err != nil {
		return nil, err
	}

	if err := checkMinValue("fd-limit", *systemdUnitParams.FdLimit, minFdLimit); err != nil {
		return nil, err
	}

	if err := setDefaults(&systemdUnitParams.StateboardFdLimit, defaultStateboardFdLimit); err != nil {
		return nil, err
	}

	if err := checkMinValue("stateboard-fd-limit", *systemdUnitParams.StateboardFdLimit, minStateboardFdLimit); err != nil {
		return nil, err
	}

	return systemdUnitParams, nil
}

func initSystemdDir(baseDirPath string, ctx *context.Ctx) error {
	log.Infof("Initialize systemd dir")

	systemdUnitParams, err := getSystemdUnitParams(ctx)
	if err != nil {
		return err
	}

	systemdCtx := getSystemdCtx(ctx, *systemdUnitParams)

	systemdFilesTemplate, err := getSystemdTemplate(ctx)
	if err != nil {
		return err
	}

	if err := systemdFilesTemplate.Instantiate(baseDirPath, systemdCtx); err != nil {
		return fmt.Errorf("Failed to instantiate systemd dir: %s", err)
	}

	return nil
}

func getSystemdTemplate(ctx *context.Ctx) (templates.Template, error) {
	var err error

	systemdFilesTemplate := systemdAppFilesTemplate

	// app unit file template
	appUnit := defaultAppUnitTemplate
	if ctx.Pack.UnitTemplatePath != "" {
		appUnit.Content, err = common.GetFileContent(ctx.Pack.UnitTemplatePath)
		if err != nil {
			return nil, fmt.Errorf("Failed to read specified unit template: %s", err)
		}
	}

	systemdFilesTemplate.AddFiles(appUnit)

	// app instantiated unit file template
	appInstUnit := defaultAppInstUnitTemplate
	if ctx.Pack.InstUnitTemplatePath != "" {
		appInstUnit.Content, err = common.GetFileContent(ctx.Pack.InstUnitTemplatePath)
		if err != nil {
			return nil, fmt.Errorf("Failed to read specified instantiated unit template: %s", err)
		}
	}

	systemdFilesTemplate.AddFiles(appInstUnit)

	// stateboard unit file template
	if ctx.Running.WithStateboard {
		stateboardUnit := defaultStateboardUnitTemplate
		if ctx.Pack.StatboardUnitTemplatePath != "" {
			stateboardUnit.Content, err = common.GetFileContent(ctx.Pack.StatboardUnitTemplatePath)
			if err != nil {
				return nil, fmt.Errorf("Failed to read specified stateboard unit template: %s", err)
			}
		}
		systemdFilesTemplate.AddFiles(stateboardUnit)
	} else {
		log.Warnf(
			"App directory doesn't contain stateboard entrypoint script `%s`. "+
				"Stateboard systemd service unit file wouldn't be delivered",
			ctx.Running.StateboardEntrypoint,
		)
	}

	return &systemdFilesTemplate, nil
}

func getSystemdCtx(ctx *context.Ctx, systemdUnitParams SystemdUnitParams) *map[string]interface{} {
	systemdCtx := make(map[string]interface{})

	systemdCtx["Name"] = ctx.Project.Name
	systemdCtx["StateboardName"] = ctx.Project.StateboardName

	systemdCtx["DefaultWorkDir"] = project.GetInstanceWorkDir(ctx, "default")
	systemdCtx["InstanceWorkDir"] = project.GetInstanceWorkDir(ctx, instanceNameSpecifier)
	systemdCtx["StateboardWorkDir"] = project.GetStateboardWorkDir(ctx)

	systemdCtx["DefaultPidFile"] = project.GetInstancePidFile(ctx, "default")
	systemdCtx["InstancePidFile"] = project.GetInstancePidFile(ctx, instanceNameSpecifier)
	systemdCtx["StateboardPidFile"] = project.GetStateboardPidFile(ctx)

	systemdCtx["DefaultConsoleSock"] = project.GetInstanceConsoleSock(ctx, "default")
	systemdCtx["InstanceConsoleSock"] = project.GetInstanceConsoleSock(ctx, instanceNameSpecifier)
	systemdCtx["StateboardConsoleSock"] = project.GetStateboardConsoleSock(ctx)

	systemdCtx["ConfPath"] = ctx.Running.ConfPath

	systemdCtx["AppEntrypointPath"] = project.GetAppEntrypointPath(ctx)
	systemdCtx["StateboardEntrypointPath"] = project.GetStateboardEntrypointPath(ctx)

	systemdCtx["FdLimit"] = systemdUnitParams.FdLimit
	systemdCtx["StateboardFdLimit"] = systemdUnitParams.StateboardFdLimit

	if ctx.Tarantool.TarantoolIsEnterprise {
		systemdCtx["Tarantool"] = filepath.Join(ctx.Running.AppDir, "tarantool")
	} else {
		systemdCtx["Tarantool"] = "/usr/bin/tarantool"
	}

	return &systemdCtx
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
Environment=TARANTOOL_CFG={{ .ConfPath }}
Environment=TARANTOOL_PID_FILE={{ .DefaultPidFile }}
Environment=TARANTOOL_CONSOLE_SOCK={{ .DefaultConsoleSock }}
Environment=TARANTOOL_NET_MSG_MAX={{ .NetMsgMax }}

LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Increase fd limit for Vinyl
LimitNOFILE={{ .FdLimit }}

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
Environment=TARANTOOL_CFG={{ .ConfPath }}
Environment=TARANTOOL_PID_FILE={{ .InstancePidFile }}
Environment=TARANTOOL_CONSOLE_SOCK={{ .InstanceConsoleSock }}
Environment=TARANTOOL_INSTANCE_NAME=%i
Environment=TARANTOOL_NET_MSG_MAX={{ .NetMsgMax }}

LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Increase fd limit for Vinyl
LimitNOFILE={{ .FdLimit }}

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
Environment=TARANTOOL_CFG={{ .ConfPath }}
Environment=TARANTOOL_PID_FILE={{ .StateboardPidFile }}
Environment=TARANTOOL_CONSOLE_SOCK={{ .StateboardConsoleSock }}
Environment=TARANTOOL_NET_MSG_MAX={{ .NetMsgMax }}

LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Increase fd limit
LimitNOFILE={{ .StateboardFdLimit }}

# Systemd waits until all xlogs are recovered
TimeoutStartSec=86400s
# Give a reasonable amount of time to close xlogs
TimeoutStopSec=10s

[Install]
WantedBy=multi-user.target
Alias={{ .StateboardName }}
`
)
