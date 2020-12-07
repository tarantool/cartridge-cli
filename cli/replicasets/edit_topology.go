package replicasets

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"

	"github.com/tarantool/cartridge-cli/cli/templates"
)

type EditReplicasetOpts struct {
	ReplicasetUUID  string
	ReplicasetAlias string

	Roles       []string
	AllRW       *bool
	Weight      *float64
	VshardGroup *string

	JoinInstancesURIs     []string
	FailoverPriorityUUIDs []string
}

type EditInstanceOpts struct {
	InstanceUUID string
	Expelled     bool
}

type EditInstancesOpts []*EditInstanceOpts

func editReplicaset(conn net.Conn, opts *EditReplicasetOpts) (*TopologyReplicaset, error) {
	replicasetInput := serializeEditReplicasetOpts(opts)

	formatTopologyReplicasetFunc, err := templates.GetTemplatedStr(
		&formatTopologyReplicasetFuncTemplate, map[string]string{
			"FormatTopologyReplicasetFuncName": formatTopologyReplicasetFuncName,
		},
	)

	if err != nil {
		return nil, project.InternalError("Failed to compute get topology replicaset function body: %s", err)
	}

	editReplicasetBody, err := templates.GetTemplatedStr(&editReplicasetBodyTemplate, map[string]string{
		"ReplicasetInput":                  replicasetInput,
		"FormatTopologyReplicasetFuncName": formatTopologyReplicasetFuncName,
		"FormatTopologyReplicasetFunc":     formatTopologyReplicasetFunc,
	})

	if err != nil {
		return nil, project.InternalError("Failed to compute edit_topology call body: %s", err)
	}

	newTopologyReplicasetRaw, err := common.EvalTarantoolConn(conn, editReplicasetBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to edit topology: %s", err)
	}

	newTopologyReplicaset, err := parseTopologyReplicaset(newTopologyReplicasetRaw)
	if err != nil {
		return nil, project.InternalError("Topology is specified in a bad format: %s", err)
	}

	return newTopologyReplicaset, nil
}

func editInstances(conn net.Conn, opts *EditInstancesOpts) (bool, error) {
	// wait for replicaset is healthy

	instancesInput := serializeEditInstancesOpts(opts)

	editInstanceBody, err := templates.GetTemplatedStr(&editInstanceBodyTemplate, map[string]string{
		"InstancesInput": instancesInput,
	})

	if err != nil {
		return false, project.InternalError("Failed to get edit topology body by template: %s", err)
	}

	_, err = common.EvalTarantoolConn(conn, editInstanceBody)
	if err != nil {
		return false, fmt.Errorf("Failed to edit topology: %s", err)
	}

	return true, nil
}

func serializeEditReplicasetOpts(opts *EditReplicasetOpts) string {
	var optsStrings []string

	appendStringOpt(&optsStrings, "uuid", &opts.ReplicasetUUID)
	appendStringOpt(&optsStrings, "alias", &opts.ReplicasetAlias)

	appendStringsSliceOpt(&optsStrings, "roles", opts.Roles)
	appendBoolOpt(&optsStrings, "all_rw", opts.AllRW)
	appendFloatOpt(&optsStrings, "weight", opts.Weight)
	appendStringOpt(&optsStrings, "vshard_group", opts.VshardGroup)

	appendJoinServersOpt(&optsStrings, "join_servers", opts.JoinInstancesURIs)
	appendStringsSliceOpt(&optsStrings, "failover_priority", opts.FailoverPriorityUUIDs)

	return strings.Join(optsStrings, ", ")
}

func serializeEditInstancesOpts(opts *EditInstancesOpts) string {
	instancesString := make([]string, len(*opts))

	for i, editInstanceOpts := range *opts {
		instancesString[i] = serializeEditInstanceOpts(editInstanceOpts)
	}

	return strings.Join(instancesString, ", ")
}

func serializeEditInstanceOpts(opts *EditInstanceOpts) string {
	var optsStrings []string

	appendStringOpt(&optsStrings, "uuid", &opts.InstanceUUID)
	appendBoolOpt(&optsStrings, "expelled", &opts.Expelled)

	return fmt.Sprintf("{ %s }", strings.Join(optsStrings, ", "))
}

func appendStringOpt(optsStrings *[]string, optName string, optValue *string) {
	if optValue == nil || *optValue == "" {
		return
	}

	optString := fmt.Sprintf("%s = '%s'", optName, *optValue)
	*optsStrings = append(*optsStrings, optString)
}

func appendBoolOpt(optsStrings *[]string, optName string, optValue *bool) {
	if optValue == nil {
		return
	}

	optString := fmt.Sprintf("%s = %t", optName, *optValue)
	*optsStrings = append(*optsStrings, optString)
}

func appendFloatOpt(optsStrings *[]string, optName string, optValue *float64) {
	if optValue == nil {
		return
	}

	formattedOpt := strconv.FormatFloat(*optValue, 'f', -1, 64)

	optString := fmt.Sprintf("%s = %s", optName, formattedOpt)
	*optsStrings = append(*optsStrings, optString)
}

func appendStringsSliceOpt(optsStrings *[]string, optName string, optValue []string) {
	if optValue == nil {
		return
	}

	optString := fmt.Sprintf("%s = %s", optName, serializeStringsSlice(optValue))
	*optsStrings = append(*optsStrings, optString)
}

func appendJoinServersOpt(optsStrings *[]string, optName string, instancesURIs []string) {
	if len(instancesURIs) == 0 {
		return
	}

	joinServerStrings := make([]string, len(instancesURIs))
	for i, instancesURI := range instancesURIs {
		joinServerStrings[i] = fmt.Sprintf("{ uri = '%s' }", instancesURI)
	}

	optString := fmt.Sprintf("%s = { %s }", optName, strings.Join(joinServerStrings, ", "))
	*optsStrings = append(*optsStrings, optString)
}

func serializeStringsSlice(stringsSlice []string) string {
	elemStrings := make([]string, len(stringsSlice))
	for i, elem := range stringsSlice {
		elemStrings[i] = fmt.Sprintf("'%s'", elem)
	}

	return fmt.Sprintf("{ %s }", strings.Join(elemStrings, ", "))
}

var (
	editReplicasetBodyTemplate = `
local cartridge = require('cartridge')

{{ .FormatTopologyReplicasetFunc }}

local res, err = cartridge.admin_edit_topology({
	replicasets = {
		{ {{ .ReplicasetInput }} },
	}
})

if err ~= nil then
	return nil, err
end

assert(#res.replicasets == 1)
local replicaset = res.replicasets[1]

local topology_replicaset = {{ .FormatTopologyReplicasetFuncName }}(replicaset)

return topology_replicaset, nil
`

	editInstanceBodyTemplate = `
local cartridge = require('cartridge')

local res, err = cartridge.admin_edit_topology({
	servers = {
		{{ .InstancesInput }},
	}
})

if err ~= nil then
	return nil, err
end

return true, nil
`

	getServersBody = `
local cartridge = require('cartridge')
return cartridge.admin_get_servers()
`
)
