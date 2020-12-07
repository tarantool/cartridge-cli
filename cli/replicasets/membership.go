package replicasets

import (
	"fmt"
	"net"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/templates"
)

type MembershipInstance struct {
	URI string

	Alias  string
	UUID   string
	Status string
}

type MembershipInstances map[string]*MembershipInstance

func connectToMembership(conn net.Conn, runningInstancesNames []string, instancesConf *InstancesConf) error {
	// probe all instances mentioned in topology
	var urisToProbe []string

	for _, instanceName := range runningInstancesNames {
		instanceConf, found := (*instancesConf)[instanceName]
		if !found {
			return fmt.Errorf("Instance %s  isn't found in instances config", instanceName)
		}

		urisToProbe = append(urisToProbe, instanceConf.URI)
	}

	probeInstancesBody, err := templates.GetTemplatedStr(&probeInstancesBodyTemplate, map[string]string{
		"URIsToProbe": serializeStringsSlice(urisToProbe),
	})

	if err != nil {
		return project.InternalError("Failed to compute probe instances function body: %s", err)
	}

	if _, err := common.EvalTarantoolConn(conn, probeInstancesBody); err != nil {
		return fmt.Errorf("Failed to probe all instances mentioned in replicasets: %s", err)
	}

	return nil
}

func getMembershipInstancesFromConn(conn net.Conn) (*MembershipInstances, error) {
	membershipInstancesRaw, err := common.EvalTarantoolConn(conn, getMembershipInstancesBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to get membership members: %s", err)
	}

	membershipInstancesRawSlice, err := common.ConvertToSlice(membershipInstancesRaw)
	if err != nil {
		return nil, project.InternalError("Membership members are returned in a bad format: %s", err)
	}

	membershipInstances := make(MembershipInstances)

	for _, instanceRaw := range membershipInstancesRawSlice {
		instanceMap, err := common.ConvertToMapWithStringKeys(instanceRaw)
		if err != nil {
			return nil, project.InternalError("Instance received in wrong format: %s", err)
		}

		instance := MembershipInstance{}

		stringFieldsMap := map[string]*string{
			"uri":    &instance.URI,
			"alias":  &instance.Alias,
			"uuid":   &instance.UUID,
			"status": &instance.Status,
		}

		for key, valuePtr := range stringFieldsMap {
			if err := getStringValueFromMap(instanceMap, key, valuePtr); err != nil {
				return nil, project.InternalError("Instance received in wrong format: %s", err)
			}
		}

		membershipInstances[instance.URI] = &instance
	}

	return &membershipInstances, nil
}

var (
	probeInstancesBodyTemplate = `
local cartridge = require('cartridge')

for _, uri in ipairs({{ .URIsToProbe }}) do
    local ok, err = cartridge.admin_probe_server(uri)
    if not ok then
		return nil, err
    end
end

return true
`

	getMembershipInstancesBody = `
local membership = require('membership')

local instances = {}

local members = membership.members()

for uri, member in pairs(members) do
	local uuid
	if member.payload ~= nil and member.payload.uuid ~= nil then
		uuid = member.payload.uuid
	end

	local alias
	if member.payload ~= nil and member.payload.alias ~= nil then
		alias = member.payload.alias
	end

	local instance = {
		uri = uri,
		alias = alias,
		uuid = uuid,
		status = member.status,
	}

	table.insert(instances, instance)
end

return instances
`
)
