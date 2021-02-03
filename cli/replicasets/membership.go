package replicasets

import (
	"fmt"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/vmihailenco/msgpack/v5"
)

type MembershipInstance struct {
	URI string

	Alias  string
	UUID   string
	Status string
}

type MembershipInstances map[string]*MembershipInstance

func (membershipInstance *MembershipInstance) DecodeMsgpack(d *msgpack.Decoder) error {
	return common.DecodeMsgpackStruct(d, membershipInstance)
}

func connectToMembership(conn *connector.Conn, runningInstancesNames []string, instancesConf *InstancesConf) error {
	// probe all instances mentioned in topology
	var urisToProbe []string

	for _, instanceName := range runningInstancesNames {
		instanceConf, found := (*instancesConf)[instanceName]
		if !found {
			return fmt.Errorf("Instance %s  isn't found in instances config", instanceName)
		}

		urisToProbe = append(urisToProbe, instanceConf.URI)
	}

	if _, err := conn.Exec(connector.EvalReq(probeInstancesBody, urisToProbe)); err != nil {
		return fmt.Errorf("Failed to probe all instances mentioned in replica sets: %s", err)
	}

	return nil
}

func getMembershipInstancesFromConn(conn *connector.Conn) (*MembershipInstances, error) {
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
