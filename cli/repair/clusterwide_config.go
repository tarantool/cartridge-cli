package repair

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/tarantool/cartridge-cli/cli/common"
	"gopkg.in/yaml.v2"
)

const (
	configDirName        = "config"
	topologyConfFilename = "topology.yml"

	keyInstances            = "servers"
	keyReplicasets          = "replicasets"
	keyInstanceAdvertiseURI = "uri"
	keyInstanceDisabled     = "disabled"

	keyInstanceReplicasetUUID = "replicaset_uuid"
	keyReplicasetLeaders      = "master"
	keyReplicasetAlias        = "alias"
	keyReplicasetRoles        = "roles"

	expelledState          = "expelled"
	unnamedReplicasetAlias = "unnamed"
)

type RawConfType map[interface{}]interface{}

type InstanceConfType struct {
	AdvertiseURI   string
	ReplicasetUUID string

	IsExpelled bool
	IsDisabled bool

	Raw RawConfType
}

type ReplicasetConfType struct {
	Alias     string
	Instances []string
	Leaders   []string
	Roles     []string

	Raw RawConfType
}

type TopologyConfType struct {
	Path    string
	Content []byte

	Raw RawConfType

	InstancesRaw RawConfType
	Instances    map[string]InstanceConfType

	ReplicasetsRaw RawConfType
	Replicasets    map[string]ReplicasetConfType
}

func getTopologyConfPath(workDir string) string {
	return filepath.Join(workDir, configDirName, topologyConfFilename)
}

func getTopologyConf(workDir string) (*TopologyConfType, error) {
	var err error
	var topologyConf TopologyConfType

	if fileInfo, err := os.Stat(workDir); err != nil {
		return nil, fmt.Errorf("Failed to use instance workdir: %s", err)
	} else if !fileInfo.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", workDir)
	}

	topologyConf.Path = getTopologyConfPath(workDir)
	if _, err := os.Stat(topologyConf.Path); err != nil {
		return nil, fmt.Errorf("Failed to use topology config path: %s", err)
	}

	topologyConf.Content, err = common.GetFileContentBytes(topologyConf.Path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read config: %s", err)
	}

	if err := yaml.Unmarshal(topologyConf.Content, &topologyConf.Raw); err != nil {
		return nil, fmt.Errorf("Failed to parse config: %s", err)
	}

	instancesConf, err := getInstancesConf(&topologyConf)
	if err != nil {
		return nil, fmt.Errorf("Failed to get instances config: %s", err)
	}
	topologyConf.Instances = *instancesConf

	replicasetsConf, err := getReplicasetsConf(&topologyConf)
	if err != nil {
		return nil, fmt.Errorf("Failed to get replicasets config: %s", err)
	}
	topologyConf.Replicasets = *replicasetsConf

	return &topologyConf, nil
}

// INSTANCES

func getInstancesConf(topologyConf *TopologyConfType) (*map[string]InstanceConfType, error) {
	instancesConfRaw, found := topologyConf.Raw[keyInstances]
	if !found {
		return nil, fmt.Errorf("Topology config doesn't contain %q key", keyInstances)
	}

	instancesConfRawMap, ok := instancesConfRaw.(RawConfType)
	if !ok {
		return nil, fmt.Errorf("%q value isn't a map", keyInstances)
	}

	topologyConf.InstancesRaw = instancesConfRawMap

	instancesConf := make(map[string]InstanceConfType)

	for instanceUUIDRaw, instanceConfRaw := range instancesConfRawMap {
		instanceUUID, ok := instanceUUIDRaw.(string)
		if !ok {
			return nil, fmt.Errorf("Instance UUID isn't a string")
		}

		var instanceConf InstanceConfType

		switch conf := instanceConfRaw.(type) {
		case string:
			if conf != expelledState {
				return nil, fmt.Errorf("Instance %s is in the unknown state %s", instanceUUID, conf)
			}
			instanceConf.IsExpelled = true
		case RawConfType:
			isDisabled, ok := conf[keyInstanceDisabled]
			if !ok {
				return nil, fmt.Errorf("Instance %s config doesn't contain %q key", instanceUUID, keyInstanceDisabled)
			}
			instanceConf.IsDisabled, ok = isDisabled.(bool)
			if !ok {
				return nil, fmt.Errorf("Instance %s has %q that isn't a bool", instanceUUID, keyInstanceDisabled)
			}

			advertiseURI, ok := conf[keyInstanceAdvertiseURI]
			if !ok {
				return nil, fmt.Errorf("Instance %s config doesn't contain %q key", instanceUUID, keyInstanceAdvertiseURI)
			}
			instanceConf.AdvertiseURI, ok = advertiseURI.(string)
			if !ok {
				return nil, fmt.Errorf("Instance %s has %q that isn't a string", instanceUUID, keyInstanceAdvertiseURI)
			}

			replicasetUUID, ok := conf[keyInstanceReplicasetUUID]
			if !ok {
				return nil, fmt.Errorf("Instance %s config doesn't contain %q key", instanceUUID, keyInstanceReplicasetUUID)
			}
			instanceConf.ReplicasetUUID, ok = replicasetUUID.(string)
			if !ok {
				return nil, fmt.Errorf("Instance %s has %q that isn't a string", instanceUUID, keyInstanceReplicasetUUID)
			}

			instanceConf.Raw = conf
		default:
			return nil, fmt.Errorf("Instance %s config isn't a map or a string", instanceUUID)
		}

		instancesConf[instanceUUID] = instanceConf
	}

	return &instancesConf, nil
}

func setInstanceURIRaw(topologyConf *TopologyConfType, instanceUUID, newURI string) error {
	instanceConf, ok := topologyConf.Instances[instanceUUID]
	if !ok {
		return fmt.Errorf("Instance %s isn't found in cluster", instanceUUID)
	}

	if instanceConf.IsExpelled {
		return fmt.Errorf("Instance %s is expelled", instanceUUID)
	}

	instanceConf.Raw[keyInstanceAdvertiseURI] = newURI

	return nil
}

func removeInstanceFromRaw(topologyConf *TopologyConfType, instanceUUID string) {
	delete(topologyConf.InstancesRaw, instanceUUID)
}

func removeReplicasetFromRaw(topologyConf *TopologyConfType, replicasetUUID string) {
	delete(topologyConf.ReplicasetsRaw, replicasetUUID)
}

// REPLICASETS

func getReplicasetsConf(topologyConf *TopologyConfType) (*map[string]ReplicasetConfType, error) {
	replicasetsConfRaw, found := topologyConf.Raw[keyReplicasets]
	if !found {
		return nil, fmt.Errorf("Topology config doesn't contain %q key", keyReplicasets)
	}

	replicasetsConfRawMap, ok := replicasetsConfRaw.(RawConfType)
	if !ok {
		return nil, fmt.Errorf("%q value isn't a map", keyReplicasets)
	}

	topologyConf.ReplicasetsRaw = replicasetsConfRawMap

	replicasetsConf := make(map[string]ReplicasetConfType)

	for replicasetUUIDRaw, replicasetConfRaw := range replicasetsConfRawMap {
		replicasetUUID, ok := replicasetUUIDRaw.(string)
		if !ok {
			return nil, fmt.Errorf("Replicaset UUID isn't a string")
		}

		var replicasetConf ReplicasetConfType

		switch conf := replicasetConfRaw.(type) {
		case RawConfType:
			replicasetConf.Raw = conf

			// alias
			aliasRaw, ok := conf[keyReplicasetAlias]
			if !ok {
				return nil, fmt.Errorf("Replicaset %s config doesn't contain %q key", replicasetUUID, keyReplicasetAlias)
			}

			alias, ok := aliasRaw.(string)
			if !ok {
				return nil, fmt.Errorf("Replicaset %q field isn't a string", keyReplicasetAlias)
			}

			if alias != unnamedReplicasetAlias {
				replicasetConf.Alias = alias
			}

			// roles
			rolesRaw, ok := conf[keyReplicasetRoles]
			if !ok {
				return nil, fmt.Errorf("Replicaset %s config doesn't contain %q key", replicasetUUID, keyReplicasetRoles)
			}

			rolesRawConf, ok := rolesRaw.(RawConfType)
			if !ok {
				return nil, fmt.Errorf("Replicaset %s config %q field isn't a map", replicasetUUID, keyReplicasetRoles)
			}

			for roleRaw := range rolesRawConf {
				role, ok := roleRaw.(string)
				if !ok {
					return nil, fmt.Errorf("Replicaset %q map key %v isn't a string", replicasetUUID, roleRaw)
				}
				replicasetConf.Roles = append(replicasetConf.Roles, role)
			}

			// leaders
			leadersRaw, ok := conf[keyReplicasetLeaders]
			if !ok {
				return nil, fmt.Errorf("Replicaset %s config doesn't contain %q key", replicasetUUID, keyReplicasetLeaders)
			}

			// XXX: old format - master is a string

			leaders, err := common.ConvertToStringsSlice(leadersRaw)
			if err != nil {
				return nil, fmt.Errorf("Replicaset %q field isn't a list of strings: %s", keyReplicasetLeaders, err)
			}

			replicasetConf.Leaders = leaders

			// instances
			replicasetConf.Instances = make([]string, 0)

			for instanceUUID, instanceConf := range topologyConf.Instances {
				if instanceConf.ReplicasetUUID == replicasetUUID {
					replicasetConf.Instances = append(replicasetConf.Instances, instanceUUID)
				}
			}

			sort.Sort(sort.StringSlice(replicasetConf.Instances))

		default:
			return nil, fmt.Errorf("Replicaset %s config isn't a map", replicasetUUID)
		}

		replicasetsConf[replicasetUUID] = replicasetConf
	}

	return &replicasetsConf, nil
}

func setReplicasetLeadersRaw(topologyConf *TopologyConfType, replicasetUUID string, leaders []string) error {
	replicasetConf, ok := topologyConf.Replicasets[replicasetUUID]
	if !ok {
		return fmt.Errorf("Replicaset %s isn't found in the cluster", replicasetUUID)
	}

	replicasetConf.Raw[keyReplicasetLeaders] = leaders

	return nil
}
