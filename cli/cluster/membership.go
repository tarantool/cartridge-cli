package cluster

import (
	"fmt"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/vmihailenco/msgpack/v5"
)

type MembershipInstance struct {
	URI string

	Alias  string
	UUID   string
	Status string
}

type MembershipInstances map[string]*MembershipInstance

type InstancesFilter func(*MembershipInstances, map[string]string)

func (membershipInstance *MembershipInstance) DecodeMsgpack(d *msgpack.Decoder) error {
	return common.DecodeMsgpackStruct(d, membershipInstance)
}

// GetRandomInstanceName returns random instance name from the map.
// Empty string is returned if the map is empty.
func GetRandomInstanceName(instances map[string]string) string {
	for _, instanceName := range instances {
		return instanceName
	}
	return ""
}

// GetMembershipInstances returns MembershipInstances for currently configured instances
// First, it connects all instances to membership (probes all running instances by one
// of them). Then, it gets all membership instances members.
// filters are an Options pattern to perform modifications of the instances container.
func GetMembershipInstances(instancesConf *InstancesConf, ctx *context.Ctx,
	filters ...InstancesFilter) (*MembershipInstances, error) {
	runningInstances := getRunningInstances(instancesConf, ctx)
	instanceName := GetRandomInstanceName(runningInstances)
	if instanceName == "" {
		return nil, fmt.Errorf("No running instances found")
	}

	conn, err := ConnectToInstance(instanceName, ctx)
	if err != nil {
		return nil, err
	}

	log.Debugf("Connect all instances to membership")

	if err := ConnectToMembership(conn, runningInstances, instancesConf); err != nil {
		return nil, fmt.Errorf("Failed to connect instances to membership: %s", err)
	}

	membershipInstances, err := GetMembershipInstancesFromConn(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to get membership instances: %s", err)
	}

	for _, filter := range filters {
		filter(membershipInstances, runningInstances)
	}

	return membershipInstances, nil
}

func ConnectToMembership(conn *connector.Conn, runningInstancesNames map[string]string,
	instancesConf *InstancesConf) error {
	// Probe all running instances mentioned in topology.
	var urisToProbe []string

	for _, instanceName := range runningInstancesNames {
		instanceConf, found := (*instancesConf)[instanceName]
		if !found {
			return fmt.Errorf("Instance %s isn't found in instances config", instanceName)
		}

		urisToProbe = append(urisToProbe, instanceConf.URI)
	}

	if _, err := conn.Exec(connector.EvalReq(probeInstancesBody, urisToProbe)); err != nil {
		return fmt.Errorf("Failed to probe all instances mentioned in replica sets: %s", err)
	}

	return nil
}

func GetMembershipInstancesFromConn(conn *connector.Conn) (*MembershipInstances, error) {
	var membershipInstancesSlice []*MembershipInstance

	req := connector.EvalReq(getMembershipInstancesBody).SetReadTimeout(SimpleOperationTimeout)
	if err := conn.ExecTyped(req, &membershipInstancesSlice); err != nil {
		return nil, fmt.Errorf("Failed to get membership members: %s", err)
	}

	membershipInstances := make(MembershipInstances)
	for _, membershipInstance := range membershipInstancesSlice {
		membershipInstances[membershipInstance.URI] = membershipInstance
	}

	return &membershipInstances, nil
}
