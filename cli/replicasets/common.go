package replicasets

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/running"
	"gopkg.in/yaml.v2"
)

type InstanceConf struct {
	URI string `yaml:"advertise_uri"`
}

type InstancesConf map[string]*InstanceConf

func getInstancesConf(ctx *context.Ctx) (*InstancesConf, error) {
	var err error

	log.Debugf("Instances configuration file is %s", ctx.Running.RunDir)

	if _, err := os.Stat(ctx.Running.ConfPath); err != nil {
		return nil, fmt.Errorf("Failed to use instances configuration file: %s", err)
	}

	fileContentBytes, err := common.GetFileContentBytes(ctx.Running.ConfPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read instances configuration file: %s", err)
	}

	var allSectionsConf InstancesConf
	if err := yaml.Unmarshal([]byte(fileContentBytes), &allSectionsConf); err != nil {
		return nil, fmt.Errorf("Failed to parse replicasets configuration file %s: %s", ctx.Replicasets.File, err)
	}

	instancesConf := make(InstancesConf)

	appInstancePrefix := fmt.Sprintf("%s.", ctx.Project.Name)
	for key, sectionConf := range allSectionsConf {
		if strings.HasPrefix(key, appInstancePrefix) {
			parts := strings.SplitN(key, ".", 2)
			instanceName := parts[1]
			instancesConf[instanceName] = sectionConf
		}
	}

	return &instancesConf, nil
}

func getControlConn(instancesConf *InstancesConf, ctx *context.Ctx, joinInstances []string) (net.Conn, error) {
	controlInstanceName, err := getJoinedInstanceName(instancesConf, ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed find some instance joined to custer")
	}

	if controlInstanceName == "" {
		if len(joinInstances) > 0 {
			controlInstanceName = joinInstances[0]
		}
	}

	if controlInstanceName == "" {
		return nil, fmt.Errorf("Failed to find joined instance")
	}

	consoleSockPath := project.GetInstanceConsoleSock(ctx, controlInstanceName)
	conn, err := common.ConnectToTarantoolSocket(consoleSockPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Tarantool instance: %s", err)
	}

	log.Debugf("Connected to %s", consoleSockPath)

	return conn, nil
}

func getJoinedInstanceName(instancesConf *InstancesConf, ctx *context.Ctx) (string, error) {
	joinedInstances, err := getJoinedInstances(instancesConf, ctx)
	if err != nil {
		return "", fmt.Errorf("Failed to get instances connected to membership: %s", err)
	}

	var joinedInstanceName string
	for instanceURI, instance := range *joinedInstances {
		if instance.UUID != "" {
			if instance.Alias == "" {
				return "", fmt.Errorf("Failed to get alias for instance %s", instanceURI)
			}

			joinedInstanceName = instance.Alias
			break
		}
	}

	return joinedInstanceName, nil
}

func getJoinedInstances(instancesConf *InstancesConf, ctx *context.Ctx) (*MembershipInstances, error) {
	runningInstancesNames := getRunningInstances(instancesConf, ctx)
	if len(runningInstancesNames) == 0 {
		return nil, fmt.Errorf("No running instances found")
	}

	instanceName := runningInstancesNames[0]
	instanceSocketPath := project.GetInstanceConsoleSock(ctx, runningInstancesNames[0])
	conn, err := common.ConnectToTarantoolSocket(instanceSocketPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to instance %s: %s", instanceName, err)
	}

	log.Debugf("Connect all replicasets instances to membership")
	if err := connectToMembership(conn, runningInstancesNames, instancesConf); err != nil {
		return nil, fmt.Errorf("Failed to connect instances to membership: %s", err)
	}

	membershipInstances, err := getMembershipInstances(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to get membership instances: %s", err)
	}

	return membershipInstances, nil
}

func getRunningInstances(instancesConf *InstancesConf, ctx *context.Ctx) []string {
	var runningInstancesNames []string
	for instanceName := range *instancesConf {
		process := running.NewInstanceProcess(ctx, instanceName)
		if process.IsRunning() {
			runningInstancesNames = append(runningInstancesNames, instanceName)
		}
	}

	return runningInstancesNames
}

func convertToMapWithStringKeys(raw interface{}) (map[string]interface{}, error) {
	rawMap, ok := raw.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("Should be a map with string keys, got %#v", raw)
	}

	mapWithStringKeys := make(map[string]interface{})

	for keyRaw, valueRaw := range rawMap {
		keyString, ok := keyRaw.(string)
		if !ok {
			return nil, fmt.Errorf("Has non string key: %#v", keyRaw)
		}

		mapWithStringKeys[keyString] = valueRaw
	}

	return mapWithStringKeys, nil
}

func convertToSlice(raw interface{}) ([]interface{}, error) {
	iterfacesSlice, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Should be a list, got %#v", raw)
	}

	return iterfacesSlice, nil
}
