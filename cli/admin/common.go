package admin

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/tarantool/cartridge-cli/cli/templates"
	"github.com/vmihailenco/msgpack/v5"

	"github.com/apex/log"
	"github.com/spf13/pflag"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

type ArgSpec struct {
	Usage string
	Type  string
}

type ArgsSpec map[string]ArgSpec

type FuncInfo struct {
	Name  string
	Usage string
	Args  ArgsSpec
}

func (funcInfo *FuncInfo) DecodeMsgpack(d *msgpack.Decoder) error {
	return common.DecodeMsgpackStruct(d, funcInfo)
}

type FuncInfos map[string]FuncInfo

func (funcInfo *FuncInfo) Format() string {
	argsUsagesMap := make(map[string]string)

	for argName, argSpec := range funcInfo.Args {
		prettyArgName := strings.ReplaceAll(argName, "_", "-")

		argDef := fmt.Sprintf("  --%s %s", prettyArgName, argSpec.Type)
		argsUsagesMap[argDef] = argSpec.Usage
	}

	argsUsageStr := common.FormatStringStringMap(argsUsagesMap)

	funcHelpMsg, err := templates.GetTemplatedStr(&funcHelpMsgTmpl, map[string]interface{}{
		"FuncInfo":  funcInfo.Usage,
		"ArgsUsage": argsUsageStr,
	})

	if err != nil {
		panic(err)
	}

	return funcHelpMsg
}

func (funcInfos *FuncInfos) FormatUsages() string {
	usagesMap := make(map[string]string)

	for funcName, funcInfo := range *funcInfos {
		usagesMap[funcName] = funcInfo.Usage
	}

	return common.FormatStringStringMap(usagesMap)
}

func checkCtx(ctx *context.Ctx) error {
	if ctx.Project.Name == "" && ctx.Admin.ConnString == "" && ctx.Admin.InstanceName == "" {
		return fmt.Errorf("Please, specify one of --name, --instance or --conn")
	}

	if ctx.Admin.InstanceName != "" && ctx.Admin.ConnString != "" {
		return fmt.Errorf("You can specify only one of --instance or --conn")
	}

	if ctx.Admin.InstanceName != "" && ctx.Project.Name == "" {
		return fmt.Errorf("Please, specify --name")
	}

	if ctx.Admin.ConnString != "" && ctx.Project.Name != "" {
		log.Warnf("--name is ignored since --conn is specified")
	}

	return nil
}

func getAvaliableConn(ctx *context.Ctx) (*connector.Conn, error) {
	var err error
	var address string

	if ctx.Admin.ConnString != "" {
		address = ctx.Admin.ConnString
	} else if ctx.Admin.InstanceName != "" {
		address = project.GetInstanceConsoleSock(ctx, ctx.Admin.InstanceName)
	}

	if address != "" {
		conn, err := connector.Connect(address, connector.Opts{})
		if err != nil {
			return nil, fmt.Errorf("Failed to connect: %s", err)
		}

		log.Debugf("Connected to %s", address)
		return conn, nil
	}

	addresses, err := getInstanceSocketPaths(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get paths of application instances sockets: %s", err)
	}

	for _, address := range addresses {
		conn, err := connector.Connect(address, connector.Opts{})
		if err != nil {
			log.Debugf("Failed to connect to %s: %s", address, err)
			continue
		}

		log.Debugf("Connected to %s", address)
		return conn, nil
	}

	return nil, fmt.Errorf("No available sockets found in: %s", ctx.Running.RunDir)
}

func getInstanceSocketPaths(ctx *context.Ctx) ([]string, error) {
	if fileInfo, err := os.Stat(ctx.Running.RunDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("Run directory %s doesn't exist", ctx.Running.RunDir)
	} else if err != nil {
		return nil, fmt.Errorf("Failed to use specified run directory: %s", err)
	} else if !fileInfo.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", ctx.Running.RunDir)
	}

	runFiles, err := ioutil.ReadDir(ctx.Running.RunDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to list the run directory: %s", err)
	}

	if len(runFiles) == 0 {
		return nil, fmt.Errorf("Run directory %s is empty", ctx.Running.RunDir)
	}

	instanceSocketPaths := []string{}

	appInstanceSocketPrefix := fmt.Sprintf("%s.", ctx.Project.Name)
	controlSocketSuffix := ".control"
	for _, runFile := range runFiles {
		runFileName := runFile.Name()
		if !strings.HasSuffix(runFileName, controlSocketSuffix) {
			continue
		}

		if !strings.HasPrefix(runFileName, appInstanceSocketPrefix) {
			continue
		}

		instanceSocketPath := filepath.Join(ctx.Running.RunDir, runFileName)
		instanceSocketPaths = append(instanceSocketPaths, instanceSocketPath)
	}

	if len(instanceSocketPaths) == 0 {
		return nil, fmt.Errorf("No instance sockets found in %s", ctx.Running.RunDir)
	}

	return instanceSocketPaths, nil
}

func getConflictingFlagNames(argsSpec ArgsSpec, flagSet *pflag.FlagSet) []string {
	if len(argsSpec) == 0 {
		return nil
	}

	// collect all defined `cartridge admin` flags
	cmdFlagNamesMap := make(map[string]bool)

	flagSet.VisitAll(func(f *pflag.Flag) {
		normalizedName := normalizeFlagName(f.Name)
		cmdFlagNamesMap[normalizedName] = true
	})

	// check argsSpec conflicting names
	conflictingFlagNames := []string{}
	for argName := range argsSpec {
		normalizedName := normalizeFlagName(argName)
		if _, found := cmdFlagNamesMap[normalizedName]; found {
			conflictingFlagNames = append(conflictingFlagNames, fmt.Sprintf("%q", argName))
		}
	}

	sort.Strings(conflictingFlagNames)

	return conflictingFlagNames
}

func normalizeFlagName(name string) string {
	return strings.ReplaceAll(name, "_", "-")
}

func getAdminFuncEvalTypedBody(adminFuncName string) (string, error) {
	funcBody, err := templates.GetTemplatedStr(&evalFuncGetResBodyTmpl, map[string]string{
		"FuncName": adminFuncName,
	})
	if err != nil {
		return "", project.InternalError("Failed to instantiate func call body template: %s", err)
	}

	return funcBody, nil
}

func getCliExtError(format string, a ...interface{}) error {
	const cliExtErrFmt = "%s. " +
		"Please update cartridge-cli-extensions module or " +
		"file an issue https://github.com/tarantool/cartridge-cli-extensions/issues/new"

	msg := fmt.Sprintf(format, a...)
	return fmt.Errorf(cliExtErrFmt, msg)
}

var (
	evalFuncGetResBodyTmpl = `
local res, err = {{ .FuncName }}(...)
assert(err == nil, err)
return res
`
)
