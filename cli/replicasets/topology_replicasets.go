package replicasets

import (
	"fmt"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/tarantool/cartridge-cli/cli/templates"
	"github.com/vmihailenco/msgpack/v5"
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

	AllRW       *bool `mapstructure:"all_rw"`
	Weight      *float64
	VshardGroup *string `mapstructure:"vshard_group"`

	Instances  TopologyInstances
	LeaderUUID string `mapstructure:"leader_uuid"`
}

type TopologyReplicasets map[string]*TopologyReplicaset

func (topologyReplicaset *TopologyReplicaset) DecodeMsgpack(d *msgpack.Decoder) error {
	return common.DecodeMsgpackStruct(d, topologyReplicaset)
}

var (
	getTopologyReplicasetsBody string
)

func init() {
	var err error

	formatTopologyReplicasetFunc, err := templates.GetTemplatedStr(
		&formatTopologyReplicasetFuncTemplate, map[string]string{
			"FormatTopologyReplicasetFuncName": formatTopologyReplicasetFuncName,
		},
	)

	if err != nil {
		panic(fmt.Errorf("Failed to compute get topology replica set function body: %s", err))
	}

	getTopologyReplicasetsBody, err = templates.GetTemplatedStr(
		&getTopologyReplicasetsBodyTemplate, map[string]string{
			"FormatTopologyReplicasetFuncName": formatTopologyReplicasetFuncName,
			"FormatTopologyReplicasetFunc":     formatTopologyReplicasetFunc,
		},
	)

	if err != nil {
		panic(fmt.Errorf("Failed to compute get topology replica set function body: %s", err))
	}
}

func (topologyReplicasets *TopologyReplicasets) GetSomeReplicaset() *TopologyReplicaset {
	for _, topologyReplicaset := range *topologyReplicasets {
		return topologyReplicaset
	}

	return nil
}

func getTopologyReplicasets(conn *connector.Conn) (*TopologyReplicasets, error) {
	req := connector.EvalReq(getTopologyReplicasetsBody).SetReadTimeout(SimpleOperationTimeout)

	var topologyReplicasetsList []*TopologyReplicaset
	if err := conn.ExecTyped(req, &topologyReplicasetsList); err != nil {
		return nil, fmt.Errorf("Failed to get current topology: %s", err)
	}

	return getTopologyReplicasetsFromList(topologyReplicasetsList), nil
}

func getTopologyReplicasetsFromList(topologyReplicasetsList []*TopologyReplicaset) *TopologyReplicasets {
	topologyReplicasets := make(TopologyReplicasets)
	for _, topologyReplicaset := range topologyReplicasetsList {
		topologyReplicasets[topologyReplicaset.UUID] = topologyReplicaset
	}

	return &topologyReplicasets
}

func (topologyReplicasets *TopologyReplicasets) GetByAlias(alias string) *TopologyReplicaset {
	for _, replicaset := range *topologyReplicasets {
		if replicaset.Alias == alias {
			return replicaset
		}
	}

	return nil
}

func getTopologyReplicaset(conn *connector.Conn, replicasetAlias string) (*TopologyReplicaset, error) {
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

const (
	formatTopologyReplicasetFuncName = "format_topology_replicaset"
)

var (
	getTopologyReplicasetsBodyTemplate = `
local cartridge = require('cartridge')

{{ .FormatTopologyReplicasetFunc }}

local topology_replicasets = {}

local replicasets, err = cartridge.admin_get_replicasets()
assert(err == nil, tostring(err))

for _, replicaset in pairs(replicasets) do
	local topology_replicaset = {{ .FormatTopologyReplicasetFuncName }}(replicaset)
	table.insert(topology_replicasets, topology_replicaset)
end

return unpack(topology_replicasets)
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
