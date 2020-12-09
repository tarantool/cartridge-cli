package replicasets

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/avast/retry-go"

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

type EditReplicasetsListOpts []*EditReplicasetOpts

type EditInstanceOpts struct {
	InstanceUUID string
	Expelled     bool
}

type EditInstancesListOpts []*EditInstanceOpts

func editReplicasetsList(conn net.Conn, opts *EditReplicasetsListOpts) (*TopologyReplicasets, error) {
	replicasetsInput, err := serializeEditReplicasetsListOpts(opts)
	if err != nil {
		return nil, err
	}

	waitForHealthy, err := healthCheckIsNeeded(conn)
	if err != nil {
		return nil, err
	}

	formatTopologyReplicasetFunc, err := templates.GetTemplatedStr(
		&formatTopologyReplicasetFuncTemplate, map[string]string{
			"FormatTopologyReplicasetFuncName": formatTopologyReplicasetFuncName,
		},
	)

	if err != nil {
		return nil, project.InternalError("Failed to compute get topology replicaset function body: %s", err)
	}

	editReplicasetsBody, err := templates.GetTemplatedStr(&editReplicasetsBodyTemplate, map[string]string{
		"ReplicasetsInput":                 replicasetsInput,
		"FormatTopologyReplicasetFuncName": formatTopologyReplicasetFuncName,
		"FormatTopologyReplicasetFunc":     formatTopologyReplicasetFunc,
	})

	if err != nil {
		return nil, project.InternalError("Failed to compute edit_topology call body: %s", err)
	}

	newTopologyReplicasetsRaw, err := common.EvalTarantoolConn(conn, editReplicasetsBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to edit topology: %s", err)
	}

	newTopologyReplicasets, err := parseTopologyReplicasets(newTopologyReplicasetsRaw)
	if err != nil {
		return nil, project.InternalError("Topology is specified in a bad format: %s", err)
	}

	if waitForHealthy {
		if err := waitForClusterIsHealthy(conn); err != nil {
			return nil, fmt.Errorf("Failed to wait for cluster to become healthy: %s", err)
		}
	}

	return newTopologyReplicasets, nil
}

func editReplicaset(conn net.Conn, opts *EditReplicasetOpts) (*TopologyReplicaset, error) {
	editReplicasetsOpts := &EditReplicasetsListOpts{opts}
	newTopologyReplicasets, err := editReplicasetsList(conn, editReplicasetsOpts)
	if err != nil {
		return nil, err
	}

	if len(*newTopologyReplicasets) != 1 {
		return nil, project.InternalError("One replicaset should be returned, got %#v", newTopologyReplicasets)
	}

	newTopologyReplicaset := newTopologyReplicasets.GetSomeReplicaset()

	return newTopologyReplicaset, nil
}

func editInstances(conn net.Conn, opts *EditInstancesListOpts) (bool, error) {
	// wait for replicaset is healthy

	instancesInput := serializeEditInstancesListOpts(opts)

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

func waitForClusterIsHealthy(conn net.Conn) error {
	retryOpts := []retry.Option{
		retry.MaxDelay(1 * time.Second),
		retry.Attempts(30),
		retry.LastErrorOnly(true),
		retry.RetryIf(func(err error) bool {
			return !strings.Contains(err.Error(), "Received in bad format")
		}),
	}

	checkClusterIsHealthyFunc := func() error {
		isHealthyRaw, err := common.EvalTarantoolConn(conn, getClusterIsHealthyBody)
		if err != nil {
			return fmt.Errorf("Failed to get replicaset status: %s", err)
		}

		isHealthy, ok := isHealthyRaw.(bool)
		if !ok {
			return project.InternalError("Received in bad format: Is healthy isn't a bool: %v", isHealthyRaw)
		}

		if !isHealthy {
			return fmt.Errorf("Cluster isn't healthy")
		}

		return nil
	}

	return retry.Do(checkClusterIsHealthyFunc, retryOpts...)
}

func serializeEditReplicasetsListOpts(opts *EditReplicasetsListOpts) (string, error) {
	replicasetsResults := make([]string, len(*opts))
	for i, replicasetOpts := range *opts {
		replicasetRes, err := getEditReplicasetOptsString(replicasetOpts)
		if err != nil {
			return "", project.InternalError("Failed to serialize edit replicaset opts: %s", err)
		}

		replicasetsResults[i] = replicasetRes
	}

	res, err := templates.GetTemplatedStr(&tableTemplate, map[string]string{
		"OptsString": strings.Join(replicasetsResults, ", "),
	})
	if err != nil {
		return "", project.InternalError("Failed to serialize edit replicasets opts: %s", err)
	}

	return res, nil
}

func getEditReplicasetOptsString(opts *EditReplicasetOpts) (string, error) {
	var optsStrings []string

	appendStringOpt(&optsStrings, "uuid", &opts.ReplicasetUUID)
	appendStringOpt(&optsStrings, "alias", &opts.ReplicasetAlias)

	appendStringsSliceOpt(&optsStrings, "roles", opts.Roles)
	appendBoolOpt(&optsStrings, "all_rw", opts.AllRW)
	appendFloatOpt(&optsStrings, "weight", opts.Weight)
	appendStringOpt(&optsStrings, "vshard_group", opts.VshardGroup)

	appendJoinServersOpt(&optsStrings, "join_servers", opts.JoinInstancesURIs)
	appendStringsSliceOpt(&optsStrings, "failover_priority", opts.FailoverPriorityUUIDs)

	res, err := templates.GetTemplatedStr(&tableTemplate, map[string]string{
		"OptsString": strings.Join(optsStrings, ", "),
	})

	if err != nil {
		return "", project.InternalError("Failed to serialize edit replicaset opts: %s", err)
	}

	return res, nil
}

func serializeEditInstancesListOpts(opts *EditInstancesListOpts) string {
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
	tableTemplate = `{ {{ .OptsString }} }`

	editReplicasetsBodyTemplate = `
local cartridge = require('cartridge')

{{ .FormatTopologyReplicasetFunc }}

local res, err = cartridge.admin_edit_topology({
	replicasets = {{ .ReplicasetsInput }}
})

if err ~= nil then
	return nil, err
end

local replicasets = res.replicasets

local topology_replicasets = {}
for _, replicaset in pairs(replicasets) do
	local topology_replicaset = {{ .FormatTopologyReplicasetFuncName }}(replicaset)
	table.insert(topology_replicasets, topology_replicaset)
end

return topology_replicasets, nil
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

	getClusterIsHealthyBody = `
local cartridge = require('cartridge')
return cartridge.is_healthy()
`
)
