package running

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

var (
	confFilePatterns = []string{
		"*.yml",
		"*.yaml",
	}
)

func CollectInstancesFromConf(ctx *context.Ctx) ([]string, error) {
	var instances []string

	// collect conf files
	var confFilePaths []string
	if fileInfo, err := os.Stat(ctx.Running.ConfPath); err != nil {
		return nil, fmt.Errorf("Failed to use conf path: %s", err)
	} else if fileInfo.IsDir() {
		for _, pattern := range confFilePatterns {
			paths, err := filepath.Glob(filepath.Join(ctx.Running.ConfPath, pattern))
			if err != nil {
				return nil, err
			}

			confFilePaths = append(confFilePaths, paths...)
		}
	} else {
		confFilePaths = append(confFilePaths, ctx.Running.ConfPath)
	}

	addedInstances := make(map[string]struct{})

	// read files
	for _, confFilePath := range confFilePaths {
		instancesMap, err := common.ParseYmlFile(confFilePath)
		if err != nil {
			return nil, fmt.Errorf("Failed to read configuration from file: %s", err)
		}

		for instanceID := range instancesMap {
			instanceIDParts := strings.SplitN(instanceID, ".", 2)
			if len(instanceIDParts) < 2 {
				continue
			}

			appName := instanceIDParts[0]
			if appName != ctx.Project.Name {
				continue
			}

			instanceName := instanceIDParts[1]
			if _, found := addedInstances[instanceName]; found {
				return nil, fmt.Errorf("Duplicate config section %s", instanceName)
			}

			instances = append(instances, instanceName)
			addedInstances[instanceName] = struct{}{}
		}
	}

	return instances, nil
}

func collectProcesses(ctx *context.Ctx) (*ProcessesSet, error) {
	processes := ProcessesSet{}

	if ctx.Running.WithStateboard {
		process := NewStateboardProcess(ctx)
		processes.Add(process)
	}

	if !ctx.Running.StateboardOnly {
		for _, instance := range ctx.Running.Instances {
			process := NewInstanceProcess(ctx, instance)
			processes.Add(process)

		}
	}

	return &processes, nil
}

func formatEnv(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

func buildNotifySocket(process *Process) error {
	var err error

	if _, err := os.Stat(process.notifySockPath); err == nil {
		if err := os.Remove(process.notifySockPath); err != nil {
			return fmt.Errorf("Failed to remove existed notify socket: %s", err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	process.notifyConn, err = net.ListenPacket("unixgram", process.notifySockPath)
	if err != nil {
		return fmt.Errorf(
			"Failed to bind socket: %s. Probably socket path exceeds UNIX_PATH_MAX limit",
			err,
		)
	}

	process.env = append(process.env, formatEnv("NOTIFY_SOCKET", process.notifySockPath))
	return nil
}
