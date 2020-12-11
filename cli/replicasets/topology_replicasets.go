package replicasets

import (
	"fmt"
	"net"
	"reflect"

	"github.com/tarantool/cartridge-cli/cli/templates"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"
)

type TopologyInstance struct {
	Alias string
	UUID  string
	URI   string

	Zone string

	Expelled bool
}

type TopologyInstances []*TopologyInstance

type TopologyReplicaset struct {
	UUID string

	Alias  string
	Status string
	Roles  []string

	AllRW       *bool
	Weight      *float64
	VshardGroup *string

	Instances  TopologyInstances
	LeaderUUID string
}

type TopologyReplicasets map[string]*TopologyReplicaset

func (topologyReplicasets *TopologyReplicasets) GetSomeReplicaset() *TopologyReplicaset {
	for _, topologyReplicaset := range *topologyReplicasets {
		return topologyReplicaset
	}

	return nil
}

func getTopologyReplicasets(conn net.Conn) (*TopologyReplicasets, error) {
	formatTopologyReplicasetFunc, err := templates.GetTemplatedStr(
		&formatTopologyReplicasetFuncTemplate, map[string]string{
			"FormatTopologyReplicasetFuncName": formatTopologyReplicasetFuncName,
		},
	)

	if err != nil {
		return nil, project.InternalError("Failed to compute get topology replica set function body: %s", err)
	}

	getTopologyReplicasetsBody, err := templates.GetTemplatedStr(
		&getTopologyReplicasetsBodyTemplate, map[string]string{
			"FormatTopologyReplicasetFuncName": formatTopologyReplicasetFuncName,
			"FormatTopologyReplicasetFunc":     formatTopologyReplicasetFunc,
		},
	)

	if err != nil {
		return nil, project.InternalError("Failed to compute get topology replica set function body: %s", err)
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

func getTopologyReplicaset(conn net.Conn, replicasetAlias string) (*TopologyReplicaset, error) {
	topologyReplicasets, err := getTopologyReplicasets(conn)
	if err != nil {
		return nil, fmt.Errorf("Failed to get current topology replica sets: %s", err)
	}

	topologyReplicaset := topologyReplicasets.GetByAlias(replicasetAlias)
	if topologyReplicaset == nil {
		return nil, fmt.Errorf("Replica set %s isn't found in current topology", replicasetAlias)
	}

	return topologyReplicaset, nil
}

func parseTopologyReplicasets(topologyReplicasetsRaw interface{}) (*TopologyReplicasets, error) {
	topologyReplicasetsRawSlice, err := common.ConvertToSlice(topologyReplicasetsRaw)
	if err != nil {
		return nil, fmt.Errorf("Replica sets received in a bad format: %s", err)
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
	replicasetMap, err := common.ConvertToMapWithStringKeys(replicasetRaw)
	if err != nil {
		return nil, fmt.Errorf("Replica set received in wrong format: %s", err)
	}

	replicaset := TopologyReplicaset{}

	stringFieldsMap := map[string]*string{
		"uuid":        &replicaset.UUID,
		"alias":       &replicaset.Alias,
		"status":      &replicaset.Status,
		"leader_uuid": &replicaset.LeaderUUID,
	}

	for key, valuePtr := range stringFieldsMap {
		if err := getStringValueFromMap(replicasetMap, key, valuePtr); err != nil {
			return nil, fmt.Errorf("Failed to get string fields: %s", err)
		}
	}

	stringPtrFieldsMap := map[string]**string{
		"vshard_group": &replicaset.VshardGroup,
	}

	for key, valuePtr := range stringPtrFieldsMap {
		if err := getStringValuePtrFromMap(replicasetMap, key, valuePtr); err != nil {
			return nil, fmt.Errorf("Failed to get string fields: %s", err)
		}
	}

	floatPtrFieldsMap := map[string]**float64{
		"weight": &replicaset.Weight,
	}

	for key, valuePtr := range floatPtrFieldsMap {
		if err := getFloatValuePtrFromMap(replicasetMap, key, valuePtr); err != nil {
			return nil, fmt.Errorf("Failed to get int fields: %s", err)
		}
	}

	boolPtrFieldsMap := map[string]**bool{
		"all_rw": &replicaset.AllRW,
	}

	for key, valuePtr := range boolPtrFieldsMap {
		if err := getBoolValuePtrFromMap(replicasetMap, key, valuePtr); err != nil {
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

	instancesRawSlice, err := common.ConvertToSlice(instancesRaw)
	if err != nil {
		return fmt.Errorf("%q should be an array, got %#v", key, instancesRaw)
	}

	topologyInstances := make(TopologyInstances, len(instancesRawSlice))

	for i, instanceRaw := range instancesRawSlice {
		instanceRawMap, err := common.ConvertToMapWithStringKeys(instanceRaw)
		if err != nil {
			return fmt.Errorf("All elements of %q should be a map, got %#v", key, instancesRaw)
		}

		topologyInstance := TopologyInstance{}

		stringFieldsMap := map[string]*string{
			"alias": &topologyInstance.Alias,
			"uuid":  &topologyInstance.UUID,
			"uri":   &topologyInstance.URI,
			"zone":  &topologyInstance.Zone,
		}

		for key, valuePtr := range stringFieldsMap {
			if err := getStringValueFromMap(instanceRawMap, key, valuePtr); err != nil {
				return fmt.Errorf("Replica set received in wrong format: %s", err)
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

func getStringValuePtrFromMap(m map[string]interface{}, key string, valuePtr **string) error {
	valueRaw, found := m[key]

	if !found {
		return nil
	}

	value, ok := valueRaw.(string)
	if !ok {
		return fmt.Errorf("%q value should be string, found %#v", key, valueRaw)
	}

	*valuePtr = &value

	return nil
}

func getFloatValuePtrFromMap(m map[string]interface{}, key string, valuePtr **float64) error {
	valueRaw, found := m[key]
	if !found {
		return nil
	}

	var value float64

	reflectValue := reflect.ValueOf(valueRaw)

	switch reflectValue.Kind() {
	case reflect.Float64, reflect.Float32:
		value = reflectValue.Float()
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8, reflect.Int:
		value = float64(reflectValue.Int())
	case reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8, reflect.Uint:
		value = float64(reflectValue.Uint())
	default:
		return fmt.Errorf("%q value should be float, found %#v", key, valueRaw)
	}

	*valuePtr = &value

	return nil
}

func getBoolValuePtrFromMap(m map[string]interface{}, key string, valuePtr **bool) error {
	valueRaw, found := m[key]
	if !found {
		return nil
	}

	value, ok := valueRaw.(bool)
	if !ok {
		return fmt.Errorf("%q value should be bool, found %#v", key, valueRaw)
	}

	*valuePtr = &value

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
			uri = server.uri,
			zone = server.zone,
		}
		table.insert(instances, instance)
	end

	local leader_uuid
	if replicaset.active_master ~= nil then
		leader_uuid = replicaset.active_master.uuid
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
		leader_uuid = leader_uuid,
	}

	return topology_replicaset
end
`
)
