package running

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"
)

var (
	confFilePatterns = []string{
		"*.yml",
		"*.yaml",
	}
)

func collectInstancesFromConf(projectCtx *project.ProjectCtx) ([]string, error) {
	var instances []string

	// collect conf files
	var confFilePaths []string
	if fileInfo, err := os.Stat(projectCtx.ConfPath); err != nil {
		return nil, fmt.Errorf("Failed to use conf path: %s", err)
	} else if fileInfo.IsDir() {
		for _, pattern := range confFilePatterns {
			paths, err := filepath.Glob(filepath.Join(projectCtx.ConfPath, pattern))
			if err != nil {
				return nil, err
			}

			confFilePaths = append(confFilePaths, paths...)
		}
	} else {
		confFilePaths = append(confFilePaths, projectCtx.ConfPath)
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
			if appName != projectCtx.Name {
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

func collectProcesses(projectCtx *project.ProjectCtx) (*ProcessesSet, error) {
	processes := ProcessesSet{}

	if projectCtx.WithStateboard {
		process := NewStateboardProcess(projectCtx)
		processes.Add(process)
	}

	if !projectCtx.StateboardOnly {
		for _, instance := range projectCtx.Instances {
			process := NewInstanceProcess(projectCtx, instance)
			processes.Add(process)
		}
	}

	return &processes, nil
}

func formatEnv(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}

func getInstancesFromArgs(args []string, projectCtx *project.ProjectCtx) ([]string, error) {
	foundInstances := make(map[string]struct{})
	var instances []string

	for _, instanceName := range args {
		if instanceName == projectCtx.Name {
			return nil, fmt.Errorf(appNameSpecifiedError)
		}

		parts := strings.SplitN(instanceName, ".", 2)

		if len(parts) > 1 {
			return nil, fmt.Errorf(instanceIDSpecified)
		}

		if instanceName != "" {
			if _, found := foundInstances[instanceName]; found {
				return nil, fmt.Errorf("Duplicate instance name: %s", instanceName)
			}

			instances = append(instances, instanceName)
			foundInstances[instanceName] = struct{}{}
		}
	}

	return instances, nil
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

const (
	appNameSpecifiedError = "Application name is specified. " +
		"Please, specify instance name(s)"
	instanceIDSpecified = `[APP_NAME].INSTANCE_NAME is specified. ` +
		"Please, specify instance name(s)"
)
