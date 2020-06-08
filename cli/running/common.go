package running

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tarantool/cartridge-cli/cli/common"
	"gopkg.in/yaml.v2"

	"github.com/tarantool/cartridge-cli/cli/project"
)

const (
	defaultLocalRunDir   = "tmp/run"
	defaultLocalDataDir  = "tmp/data"
	defaultLocalLogDir   = "tmp/logs"
	defaultLocalConfPath = "instances.yml"
)

var (
	confFilePatterns = []string{
		"*.yml",
		"*.yaml",
	}
)

func SetLocalRunningPaths(projectCtx *project.ProjectCtx) error {
	curDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Failed to get current directory: %s", err)
	}

	if projectCtx.RunDir == "" {
		projectCtx.RunDir = filepath.Join(curDir, defaultLocalRunDir)
	}
	if projectCtx.RunDir, err = filepath.Abs(projectCtx.RunDir); err != nil {
		return fmt.Errorf("Failed to get run dir absolute path: %s", err)
	}

	if projectCtx.DataDir == "" {
		projectCtx.DataDir = filepath.Join(curDir, defaultLocalDataDir)
	}
	if projectCtx.DataDir, err = filepath.Abs(projectCtx.DataDir); err != nil {
		return fmt.Errorf("Failed to get data dir absolute path: %s", err)
	}

	if projectCtx.LogDir == "" {
		projectCtx.LogDir = filepath.Join(curDir, defaultLocalLogDir)
	}
	if projectCtx.LogDir, err = filepath.Abs(projectCtx.LogDir); err != nil {
		return fmt.Errorf("Failed to get log dir absolute path: %s", err)
	}

	if projectCtx.ConfPath == "" {
		projectCtx.ConfPath = filepath.Join(curDir, defaultLocalConfPath)
	}
	if projectCtx.ConfPath, err = filepath.Abs(projectCtx.ConfPath); err != nil {
		return fmt.Errorf("Failed to get conf path absolute path: %s", err)
	}

	return nil
}

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
		conf, err := common.GetFileContent(confFilePath)
		if err != nil {
			return nil, fmt.Errorf("Failed to read configuration from file: %s", err)
		}

		instancesMap := make(map[string]interface{})
		if err := yaml.Unmarshal([]byte(conf), instancesMap); err != nil {
			return nil, fmt.Errorf("Failed to parse %s: %s", confFilePath, err)
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

	for _, instance := range projectCtx.Instances {
		process := NewInstanceProcess(projectCtx, instance)
		processes.Add(process)
	}

	return &processes, nil
}

func formatEnv(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}
