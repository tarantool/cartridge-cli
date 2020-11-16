package admin

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/apex/log"
	"github.com/spf13/pflag"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

type ArgSpec struct {
	Usage string
	Type  string
}

type ArgsSpec map[string]ArgSpec

func getAvaliableConn(ctx *context.Ctx) (net.Conn, error) {
	if err := project.SetSystemRunningPaths(ctx); err != nil {
		return nil, fmt.Errorf("Failed to get default paths: %s", err)
	}

	log.Debugf("Run directory is set to: %s", ctx.Running.RunDir)

	// Use socket of specified instance
	if ctx.Admin.InstanceName != "" {
		instanceSocketPath := project.GetInstanceConsoleSock(ctx, ctx.Admin.InstanceName)

		conn, err := common.ConnectToTarantoolSocket(instanceSocketPath)
		if err != nil {
			return nil, fmt.Errorf("Failed to use %q: %s", instanceSocketPath, err)
		}

		log.Debugf("Connected to %q", instanceSocketPath)

		return conn, nil
	}

	// find first available socket
	instanceSocketPaths, err := getInstanceSocketPaths(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get application instances sockets paths: %s", err)
	}

	for _, instanceSocketPath := range instanceSocketPaths {
		conn, err := common.ConnectToTarantoolSocket(instanceSocketPath)
		if err == nil {
			log.Debugf("Connected to %q", instanceSocketPath)

			return conn, nil
		}

		log.Debugf("Failed to use %q: %s", instanceSocketPath, err)
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

func getArgsSpec(helpResRawMap map[interface{}]interface{}) (ArgsSpec, error) {
	argsMapRaw, found := helpResRawMap["args"]
	if !found {
		return nil, nil
	}

	argsSpec := make(ArgsSpec)

	argsMap, err := convertToMap(argsMapRaw)
	if err != nil {
		return nil, fmt.Errorf("Failed to convert to map: %s", err)
	}

	for argNameRaw, argOptsMapRaw := range argsMap {
		argName, ok := argNameRaw.(string)
		if !ok {
			return nil, fmt.Errorf("Argument name isn't a string: %#v", argNameRaw)
		}

		argOptsMap, err := convertToMap(argOptsMapRaw)
		if err != nil {
			return nil, fmt.Errorf("Failed to convert %q argument opts to map: %s", argName, err)
		}

		argUsage, err := getStrValueFromRawMap(argOptsMap, "usage")
		if err != nil {
			return nil, fmt.Errorf("Failed to get argument usage: %s", err)
		}

		argType, err := getStrValueFromRawMap(argOptsMap, "type")
		if err != nil {
			return nil, fmt.Errorf("Failed to get argument type: %s", err)
		}

		argSpec := ArgSpec{
			Usage: argUsage,
			Type:  argType,
		}

		argsSpec[argName] = argSpec
	}

	return argsSpec, nil
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

func getCliExtError(format string, a ...interface{}) error {
	const cliExtErrFmt = "%s. " +
		"Please update cartridge-cli-extensions module or " +
		"file an issue https://github.com/tarantool/cartridge-cli-extensions/issues/new"

	msg := fmt.Sprintf(format, a...)
	return fmt.Errorf(cliExtErrFmt, msg)
}
