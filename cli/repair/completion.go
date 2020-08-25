package repair

import (
	"fmt"
	"net"
	"sort"

	"github.com/tarantool/cartridge-cli/cli/context"
)

func GetAllInstanceUUIDsComp(ctx *context.Ctx) ([]string, error) {
	instanceNames, err := getAppInstanceNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get application instances working directories: %s", err)
	}

	appConfigs, err := getAppConfigs(instanceNames, ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get application cluster-wide configs: %s", err)
	}

	if len(appConfigs.hashes) == 0 {
		return nil, fmt.Errorf("No application configs found")
	}

	var instanceUUIDs []string
	addedInstances := make(map[string]bool)

	for _, topologyConf := range appConfigs.confByHash {
		for instanceUUID := range topologyConf.Instances {
			if _, found := addedInstances[instanceUUID]; !found {
				addedInstances[instanceUUID] = true
				instanceUUIDs = append(instanceUUIDs, instanceUUID)
			}
		}
	}

	sort.Sort(sort.StringSlice(instanceUUIDs))

	return instanceUUIDs, nil
}

func GetInstanceHostsComp(instanceUUID string, ctx *context.Ctx) ([]string, error) {
	instanceNames, err := getAppInstanceNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get application instances working directories: %s", err)
	}

	appConfigs, err := getAppConfigs(instanceNames, ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get application cluster-wide configs: %s", err)
	}

	if len(appConfigs.hashes) == 0 {
		return nil, fmt.Errorf("No application configs found")
	}

	var instanceHosts []string
	addedHosts := make(map[string]bool)

	for _, topologyConf := range appConfigs.confByHash {
		instanceConf, found := topologyConf.Instances[instanceUUID]
		if !found {
			continue
		}
		if instanceConf.IsExpelled {
			continue
		}

		advertiseURI := instanceConf.AdvertiseURI
		instanceHost, _, err := net.SplitHostPort(advertiseURI)
		if err != nil {
			continue
		}

		if _, found := addedHosts[instanceHost]; !found {
			addedHosts[instanceHost] = true
			instanceHosts = append(instanceHosts, instanceHost)
		}

	}

	sort.Sort(sort.StringSlice(instanceHosts))

	return instanceHosts, nil
}

func GetAllReplicasetUUIDsComp(ctx *context.Ctx) ([]string, error) {
	instanceNames, err := getAppInstanceNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get application instances working directories: %s", err)
	}

	appConfigs, err := getAppConfigs(instanceNames, ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get application cluster-wide configs: %s", err)
	}

	if len(appConfigs.hashes) == 0 {
		return nil, fmt.Errorf("No application configs found")
	}

	var replicasetUUIDs []string
	addedReplicasets := make(map[string]bool)

	for _, topologyConf := range appConfigs.confByHash {
		for replicasetUUID := range topologyConf.Replicasets {
			if _, found := addedReplicasets[replicasetUUID]; !found {
				addedReplicasets[replicasetUUID] = true
				replicasetUUIDs = append(replicasetUUIDs, replicasetUUID)
			}
		}
	}

	sort.Sort(sort.StringSlice(replicasetUUIDs))

	return replicasetUUIDs, nil
}

func GetReplicasetInstancesComp(replicasetUUID string, ctx *context.Ctx) ([]string, error) {
	instanceNames, err := getAppInstanceNames(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get application instances working directories: %s", err)
	}

	appConfigs, err := getAppConfigs(instanceNames, ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get application cluster-wide configs: %s", err)
	}

	if len(appConfigs.hashes) == 0 {
		return nil, fmt.Errorf("No application configs found")
	}

	var instanceUUIDs []string
	addedInstances := make(map[string]bool)

	for _, topologyConf := range appConfigs.confByHash {
		for instanceUUID, instanceConf := range topologyConf.Instances {
			if instanceConf.ReplicasetUUID != replicasetUUID {
				continue
			}

			if _, found := addedInstances[instanceUUID]; !found {
				addedInstances[instanceUUID] = true
				instanceUUIDs = append(instanceUUIDs, instanceUUID)
			}
		}
	}

	sort.Sort(sort.StringSlice(instanceUUIDs))

	return instanceUUIDs, nil
}
