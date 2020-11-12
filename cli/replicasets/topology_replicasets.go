package replicasets

import (
	"fmt"
	"net"

	"github.com/tarantool/cartridge-cli/cli/templates"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"
)

type TopologyInstance struct {
	Alias    string
	UUID     string
	Expelled bool
}

type TopologyInstances []*TopologyInstance

type TopologyReplicaset struct {
	UUID string

	Alias  string
	Status string
	Roles  []string

	AllRW       bool
	Weight      float64
	VshardGroup string

	Instances TopologyInstances
}

type TopologyReplicasets map[string]*TopologyReplicaset

func getTopologyReplicasets(conn net.Conn) (*TopologyReplicasets, error) {
	formatTopologyReplicasetFunc, err := templates.GetTemplatedStr(
		&formatTopologyReplicasetFuncTemplate, map[string]string{
			"FormatTopologyReplicasetFuncName": formatTopologyReplicasetFuncName,
		},
	)

	if err != nil {
		return nil, project.InternalError("Failed to compute get topology replicaset function body: %s", err)
	}

	getTopologyReplicasetsBody, err := templates.GetTemplatedStr(
		&getTopologyReplicasetsBodyTemplate, map[string]string{
			"FormatTopologyReplicasetFuncName": formatTopologyReplicasetFuncName,
			"FormatTopologyReplicasetFunc":     formatTopologyReplicasetFunc,
		},
	)

	if err != nil {
		return nil, project.InternalError("Failed to compute get topology replicaset function body: %s", err)
	}

	topologyReplicasetsRaw, err := common.EvalTarantoolConn(conn, getTopologyReplicasetsBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to get current topology: %s", err)
	}

	topologyReplicasets, err := parseTopologyReplicasets(topologyReplicasetsRaw)
	if err != nil {
		return nil, project.InternalError("Topology is specified in a bad format: %s", err)
	}

	return topologyReplicasets, nil

}

func (topologyReplicasets *TopologyReplicasets) GetByAlias(alias string) *TopologyReplicaset {
	for _, replicaset := range *topologyReplicasets {
		if replicaset.Alias == alias {
			return replicaset
		}
	}

	return nil
}

func parseTopologyReplicasets(topologyReplicasetsRaw interface{}) (*TopologyReplicasets, error) {
	topologyReplicasetsRawSlice, err := convertToSlice(topologyReplicasetsRaw)
	if err != nil {
		return nil, fmt.Errorf("Replicasets received in a bad format: %s", err)
	}

	topologyReplicasets := make(TopologyReplicasets)

	for _, replicasetRaw := range topologyReplicasetsRawSlice {
		replicaset, err := parseTopologyReplicaset(replicasetRaw)
		if err != nil {
			return nil, err
		}

		topologyReplicasets[replicaset.UUID] = replicaset
	}

	return &topologyReplicasets, nil
}

func parseTopologyReplicaset(replicasetRaw interface{}) (*TopologyReplicaset, error) {
	replicasetMap, err := convertToMapWithStringKeys(replicasetRaw)
	if err != nil {
		return nil, fmt.Errorf("Replicaset received in wrong format: %s", err)
	}

	replicaset := TopologyReplicaset{}

	stringFieldsMap := map[string]*string{
		"uuid":         &replicaset.UUID,
		"alias":        &replicaset.Alias,
		"status":       &replicaset.Status,
		"vshard_group": &replicaset.VshardGroup,
	}

	for key, valuePtr := range stringFieldsMap {
		if err := getStringValueFromMap(replicasetMap, key, valuePtr); err != nil {
			return nil, fmt.Errorf("Failed to get string fields: %s", err)
		}
	}

	floatFieldsMap := map[string]*float64{
		"weight": &replicaset.Weight,
	}

	for key, valuePtr := range floatFieldsMap {
		if err := getFloatValueFromMap(replicasetMap, key, valuePtr); err != nil {
			return nil, fmt.Errorf("Failed to get int fields: %s", err)
		}
	}

	boolFieldsMap := map[string]*bool{
		"all_rw": &replicaset.AllRW,
	}

	for key, valuePtr := range boolFieldsMap {
		if err := getBoolValueFromMap(replicasetMap, key, valuePtr); err != nil {
			return nil, fmt.Errorf("Failed to get bool fields: %s", err)
		}
	}

	stringSliceFieldsMap := map[string]*[]string{
		"roles": &replicaset.Roles,
	}

	for key, valuePtr := range stringSliceFieldsMap {
		if err := getStringSliceValueFromMap(replicasetMap, key, valuePtr); err != nil {
			return nil, fmt.Errorf("Failed to get string array fields: %s", err)
		}
	}

	if err := getTopologyInstancesFromMap(replicasetMap, "instances", &replicaset.Instances); err != nil {
		return nil, fmt.Errorf("Failed to get string array fields: %s", err)
	}

	return &replicaset, nil
}

func getTopologyInstancesFromMap(m map[string]interface{}, key string, valuePtr *TopologyInstances) error {
	instancesRaw, found := m[key]
	if !found {
		return nil
	}

	instancesRawSlice, err := convertToSlice(instancesRaw)
	if err != nil {
		return fmt.Errorf("%q should be an array, got %#v", key, instancesRaw)
	}

	topologyInstances := make(TopologyInstances, len(instancesRawSlice))

	for i, instanceRaw := range instancesRawSlice {
		instanceRawMap, err := convertToMapWithStringKeys(instanceRaw)
		if err != nil {
			return fmt.Errorf("All elements of %q should be a map, got %#v", key, instancesRaw)
		}

		topologyInstance := TopologyInstance{}

		stringFieldsMap := map[string]*string{
			"alias": &topologyInstance.Alias,
			"uuid":  &topologyInstance.UUID,
		}

		for key, valuePtr := range stringFieldsMap {
			if err := getStringValueFromMap(instanceRawMap, key, valuePtr); err != nil {
				return fmt.Errorf("Replicaset received in wrong format: %s", err)
			}
		}

		topologyInstances[i] = &topologyInstance
	}

	*valuePtr = topologyInstances

	return nil
}

func getStringValueFromMap(m map[string]interface{}, key string, valuePtr *string) error {
	valueRaw, found := m[key]
	if !found {
		return nil
	}

	value, ok := valueRaw.(string)
	if !ok {
		return fmt.Errorf("%q value should be string, found %#v", key, valueRaw)
	}

	*valuePtr = value

	return nil
}

func getFloatValueFromMap(m map[string]interface{}, key string, valuePtr *float64) error {
	valueRaw, found := m[key]
	if !found {
		return nil
	}

	var value float64

	switch i := valueRaw.(type) {
	case float64:
		value = i
	case float32:
		value = float64(i)
	case int64:
		value = float64(i)
	case int32:
		value = float64(i)
	case int:
		value = float64(i)
	case uint64:
		value = float64(i)
	case uint32:
		value = float64(i)
	case uint:
		value = float64(i)
	default:
		return fmt.Errorf("%q value should be float, found %#v", key, valueRaw)
	}

	*valuePtr = value

	return nil
}

func getBoolValueFromMap(m map[string]interface{}, key string, valuePtr *bool) error {
	valueRaw, found := m[key]
	if !found {
		return nil
	}

	value, ok := valueRaw.(bool)
	if !ok {
		return fmt.Errorf("%q value should be bool, found %#v", key, valueRaw)
	}

	*valuePtr = value

	return nil
}

func getStringSliceValueFromMap(m map[string]interface{}, key string, valuePtr *[]string) error {
	valueRaw, found := m[key]
	if !found {
		return nil
	}

	value, err := common.ConvertToStringsSlice(valueRaw)
	if err != nil {
		return fmt.Errorf("%q is not an array of strings: %s", key, err)
	}

	*valuePtr = value

	return nil
}

const (
	formatTopologyReplicasetFuncName = "format_topology_replicaset"
)

var (
	getTopologyReplicasetsBodyTemplate = `
local cartridge = require('cartridge')

{{ .FormatTopologyReplicasetFunc }}

local topology_replicasets = {}

local replicasets, err = cartridge.admin_get_replicasets()
if err ~= nil then
	return nil, err
end

for _, replicaset in pairs(replicasets) do
	local topology_replicaset = {{ .FormatTopologyReplicasetFuncName }}(replicaset)
	table.insert(topology_replicasets, topology_replicaset)
end

return topology_replicasets
`

	formatTopologyReplicasetFuncTemplate = `
local function {{ .FormatTopologyReplicasetFuncName }}(replicaset)
	local instances = {}
	for _, server in pairs(replicaset.servers) do
		local instance = {
			alias = server.alias,
			uuid = server.uuid,
		}
		table.insert(instances, instance)
	end

	local topology_replicaset = {
		uuid = replicaset.uuid,
		alias = replicaset.alias,
		status = replicaset.status,
		roles = replicaset.roles,
		all_rw = replicaset.all_rw,
		weight = replicaset.weight,
		vshard_group = replicaset.vshard_group,
		instances = instances,
	}

	return topology_replicaset
end
`
)
