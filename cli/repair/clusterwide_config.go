package repair

import (
	"fmt"
	"os"
	"path/filepath"

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

type TopologyConfType struct {
	Path string

	rawConf RawConfType

	Instances    map[string]InstanceConfType
	instancesRaw RawConfType
}

// TOPOLOGY

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

	confContent, err := common.GetFileContentBytes(topologyConf.Path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read config: %s", err)
	}

	if err := yaml.Unmarshal(confContent, &topologyConf.rawConf); err != nil {
		return nil, fmt.Errorf("Failed to parse config: %s", err)
	}

	if err := topologyConf.setInstancesConf(); err != nil {
		return nil, fmt.Errorf("Failed to parse instances config: %s", err)
	}

	return &topologyConf, nil
}

func (topologyConf *TopologyConfType) setInstancesConf() error {
	instancesConfRaw, found := topologyConf.rawConf[keyInstances]
	if !found {
		return fmt.Errorf("Topology config doesn't contain %q key", keyInstances)
	}

	instancesConfRawMap, ok := instancesConfRaw.(RawConfType)
	if !ok {
		return fmt.Errorf("%q value isn't a map", keyInstances)
	}

	topologyConf.instancesRaw = instancesConfRawMap
	topologyConf.Instances = make(map[string]InstanceConfType)

	for instanceUUIDRaw, instanceConfRaw := range instancesConfRawMap {
		instanceUUID, ok := instanceUUIDRaw.(string)
		if !ok {
			return fmt.Errorf("Instance UUID isn't a string")
		}

		var instanceConf InstanceConfType

		switch conf := instanceConfRaw.(type) {
		case string:
			if conf != expelledState {
				return fmt.Errorf("Instance %s is in the unknown state %s", instanceUUID, conf)
			}
			instanceConf.IsExpelled = true
		case RawConfType:
			isDisabled, ok := conf[keyInstanceDisabled]
			if !ok {
				return fmt.Errorf("Instance %s config doesn't contain %q key", instanceUUID, keyInstanceDisabled)
			}
			instanceConf.IsDisabled, ok = isDisabled.(bool)
			if !ok {
				return fmt.Errorf("Instance %s has %q that isn't a bool", instanceUUID, keyInstanceDisabled)
			}

			advertiseURI, ok := conf[keyInstanceAdvertiseURI]
			if !ok {
				return fmt.Errorf("Instance %s config doesn't contain %q key", instanceUUID, keyInstanceAdvertiseURI)
			}
			instanceConf.AdvertiseURI, ok = advertiseURI.(string)
			if !ok {
				return fmt.Errorf("Instance %s has %q that isn't a string", instanceUUID, keyInstanceAdvertiseURI)
			}

			replicasetUUID, ok := conf[keyInstanceReplicasetUUID]
			if !ok {
				return fmt.Errorf("Instance %s config doesn't contain %q key", instanceUUID, keyInstanceReplicasetUUID)
			}
			instanceConf.ReplicasetUUID, ok = replicasetUUID.(string)
			if !ok {
				return fmt.Errorf("Instance %s has %q that isn't a string", instanceUUID, keyInstanceReplicasetUUID)
			}

			instanceConf.Raw = conf
		default:
			return fmt.Errorf("Instance %s config isn't a map or a string", instanceUUID)
		}

		topologyConf.Instances[instanceUUID] = instanceConf
	}

	return nil
}

func (topologyConf *TopologyConfType) MarshalContent() ([]byte, error) {
	content, err := yaml.Marshal(topologyConf.rawConf)
	if err != nil {
		return nil, fmt.Errorf("Failed to YAML encode: %s", err)
	}

	return content, nil
}

// INSTANCES

func (topologyConf *TopologyConfType) SetInstanceURI(instanceUUID, newURI string) error {
	instanceConf, ok := topologyConf.Instances[instanceUUID]
	if !ok {
		return fmt.Errorf("Instance %s isn't found in cluster", instanceUUID)
	}

	if instanceConf.IsExpelled {
		return fmt.Errorf("Instance %s is expelled", instanceUUID)
	}

	instanceConf.AdvertiseURI = newURI
	instanceConf.Raw[keyInstanceAdvertiseURI] = newURI

	return nil
}
