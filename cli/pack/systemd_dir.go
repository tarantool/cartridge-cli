package pack

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	minNetMsgMax = 2

	defaultInstanceFdLimit = 65535
	defaultStateboardFdLimit = 65535
	defaultNetMsgMax = 768

	tarantoolEnvKeyPrefix = "TARANTOOL_"
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

type UnitEnvArgs map[string]interface{}

type SystemdUnitParams struct {
	FdLimit           *int `yaml:"fd-limit"`
	StateboardFdLimit *int `yaml:"stateboard-fd-limit"`

	InstanceEnv   UnitEnvArgs `yaml:"instance-env"`
	StateboardEnv UnitEnvArgs `yaml:"stateboard-env"`
}

type systemdCtxParam struct {
	ArgName      string
	CtxKey       string
	DefaultValue string
	EnvArgs      UnitEnvArgs
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

	systemdCtx, err := getSystemdCtx(ctx, systemdUnitParams)
	if err != nil {
		return err
	}

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

func setDefaultTarantoolEnvValues(ctx *context.Ctx, systemdCtx *map[string]interface{}) {
	(*systemdCtx)["UnitEnv"] = map[string]interface{}{
		"TARANTOOL_APP_NAME":     ctx.Project.Name,
		"TARANTOOL_WORKDIR":      project.GetInstanceWorkDir(ctx, "default"),
		"TARANTOOL_CFG":          ctx.Running.ConfPath,
		"TARANTOOL_PID_FILE":     project.GetInstancePidFile(ctx, "default"),
		"TARANTOOL_CONSOLE_SOCK": project.GetInstanceConsoleSock(ctx, "default"),
		"TARANTOOL_NET_MSG_MAX":  defaultNetMsgMax,
	}

	(*systemdCtx)["InstUnitEnv"] = map[string]interface{}{
		"TARANTOOL_APP_NAME":      ctx.Project.Name,
		"TARANTOOL_WORKDIR":       project.GetInstanceWorkDir(ctx, instanceNameSpecifier),
		"TARANTOOL_CFG":           ctx.Running.ConfPath,
		"TARANTOOL_PID_FILE":      project.GetInstancePidFile(ctx, instanceNameSpecifier),
		"TARANTOOL_CONSOLE_SOCK":  project.GetInstanceConsoleSock(ctx, instanceNameSpecifier),
		"TARANTOOL_NET_MSG_MAX":   defaultNetMsgMax,
		"TARANTOOL_INSTANCE_NAME": instanceNameSpecifier,
	}

	(*systemdCtx)["StateboardUnitEnv"] = map[string]interface{}{
		"TARANTOOL_APP_NAME":     ctx.Project.StateboardName,
		"TARANTOOL_WORKDIR":      project.GetStateboardWorkDir(ctx),
		"TARANTOOL_CFG":          ctx.Running.ConfPath,
		"TARANTOOL_PID_FILE":     project.GetStateboardPidFile(ctx),
		"TARANTOOL_CONSOLE_SOCK": project.GetStateboardConsoleSock(ctx),
		"TARANTOOL_NET_MSG_MAX":  defaultNetMsgMax,
	}
}

func generateTarantoolEnvKey(key string) string {
	if strings.HasPrefix(key, tarantoolEnvKeyPrefix) {
		return key
	}

	formattedKey := strings.ToUpper(strings.ReplaceAll(key, "-", "_"))

	return strings.Join([]string{tarantoolEnvKeyPrefix, formattedKey}, "")
}

func checkUnitEnvIntParam(envArgs UnitEnvArgs, argName string) error {
	if value, ok := envArgs[argName]; ok {
		netMsgMax, ok := value.(int)
		if !ok {
			return fmt.Errorf("%s parameter type should be integer", argName)
		}

		if err := checkMinValue(argName, netMsgMax, minNetMsgMax); err != nil {
			return err
		}
	}

	return nil
}

func checkNetMsgMaxValue(envArgs UnitEnvArgs) error {
	if err := checkUnitEnvIntParam(envArgs, "net-msg-max"); err != nil {
		return err
	}

	if err := checkUnitEnvIntParam(envArgs, "TARANTOOL_NET_MSG_MAX"); err != nil {
		return err
	}

	return nil
}

func updateUnitEnvBySpecifiedArgs(unitEnv interface{}, envArgs UnitEnvArgs) error {
	mapUnitEnv, ok := unitEnv.(map[string]interface{})
	if !ok {
		return project.InternalError("Setting env values: can't convert (type interface {}) to type map[string]interface{}")
	}

	if err := checkNetMsgMaxValue(envArgs); err != nil {
		return err
	}

	for key, value := range envArgs {
		tarantoolEnvKey := generateTarantoolEnvKey(key)
		mapUnitEnv[tarantoolEnvKey] = value
	}

	return nil
}

func setTarantoolEnvValues(ctx *context.Ctx, systemdCtx *map[string]interface{}, systemdUnitParams *SystemdUnitParams) error {
	setDefaultTarantoolEnvValues(ctx, systemdCtx)

	if err := updateUnitEnvBySpecifiedArgs((*systemdCtx)["UnitEnv"], (*systemdUnitParams).InstanceEnv); err != nil {
		return err
	}

	if err := updateUnitEnvBySpecifiedArgs((*systemdCtx)["InstUnitEnv"], (*systemdUnitParams).InstanceEnv); err != nil {
		return err
	}

	if err := updateUnitEnvBySpecifiedArgs((*systemdCtx)["StateboardUnitEnv"], (*systemdUnitParams).StateboardEnv); err != nil {
		return err
	}

	return nil
}

func getUnitEnvStringValue(envArgs UnitEnvArgs, key string) (string, error){
	if value, ok := envArgs[key]; ok {
		result, ok := value.(string)
		if !ok {
			return "", fmt.Errorf("%s parameter type should be string", key)
		}
		return result, nil
	}

	return "", nil
}

func getSpecifiedStringArg(defaultValue string, envArgs UnitEnvArgs, argName string) (string, error){
	if arg, err := getUnitEnvStringValue(envArgs, argName); err != nil || arg != "" {
		return arg, err
	}

	tarantoolEnvKey := generateTarantoolEnvKey(argName)
	if arg, err := getUnitEnvStringValue(envArgs, tarantoolEnvKey); err != nil || arg != "" {
		return arg, err
	}

	return defaultValue, nil
}

func getSystemdCtx(ctx *context.Ctx, systemdUnitParams *SystemdUnitParams) (*map[string]interface{}, error) {
	var err error

	systemdCtx := make(map[string]interface{})

	systemdCtxParams := []systemdCtxParam{
		{
			ArgName: "app-name",
			CtxKey: "Name",
			DefaultValue: ctx.Project.Name,
			EnvArgs: (*systemdUnitParams).InstanceEnv,
		},
		{
			ArgName: "app-name",
			CtxKey: "StateboardName",
			DefaultValue: ctx.Project.StateboardName,
			EnvArgs: (*systemdUnitParams).StateboardEnv,
		},
		{
			ArgName: "workdir",
			CtxKey: "DefaultWorkDir",
			DefaultValue: project.GetInstanceWorkDir(ctx, "default"),
			EnvArgs: (*systemdUnitParams).InstanceEnv,
		},
		{
			ArgName: "workdir",
			CtxKey: "InstanceWorkDir",
			DefaultValue: project.GetInstanceWorkDir(ctx, instanceNameSpecifier),
			EnvArgs: (*systemdUnitParams).InstanceEnv,
		},
		{
			ArgName: "workdir",
			CtxKey: "StateboardWorkDir",
			DefaultValue: project.GetStateboardWorkDir(ctx),
			EnvArgs: (*systemdUnitParams).StateboardEnv,
		},
	}

	for _, param := range systemdCtxParams {
		systemdCtx[param.CtxKey], err = getSpecifiedStringArg(param.DefaultValue, param.EnvArgs, param.ArgName)
		if err != nil {
			return nil, err
		}
	}

	systemdCtx["AppEntrypointPath"] = project.GetAppEntrypointPath(ctx)
	systemdCtx["StateboardEntrypointPath"] = project.GetStateboardEntrypointPath(ctx)

	systemdCtx["FdLimit"] = systemdUnitParams.FdLimit
	systemdCtx["StateboardFdLimit"] = systemdUnitParams.StateboardFdLimit

	if ctx.Tarantool.TarantoolIsEnterprise {
		systemdCtx["Tarantool"] = filepath.Join(ctx.Running.AppDir, "tarantool")
	} else {
		systemdCtx["Tarantool"] = "/usr/bin/tarantool"
	}

	err = setTarantoolEnvValues(ctx, &systemdCtx, systemdUnitParams)
	if err != nil {
		return nil, err
	}

	return &systemdCtx, nil
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

{{ range $tarantoolEnvKey, $tarantoolEnvValue := .UnitEnv }}Environment={{ $tarantoolEnvKey }}={{ $tarantoolEnvValue }}
{{ end }}

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

{{ range $tarantoolEnvKey, $tarantoolEnvValue := .InstUnitEnv }}Environment={{ $tarantoolEnvKey }}={{ $tarantoolEnvValue }}
{{ end }}

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

{{ range $tarantoolEnvKey, $tarantoolEnvValue := .StateboardUnitEnv }}Environment={{ $tarantoolEnvKey }}={{ $tarantoolEnvValue }}
{{ end }}

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
