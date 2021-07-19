package cluster

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/running"
	"github.com/vmihailenco/msgpack/v5"
	"gopkg.in/yaml.v2"
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

type InstanceConf struct {
	URI string `yaml:"advertise_uri"`
}

type InstancesConf map[string]*InstanceConf

const (
	SimpleOperationTimeout = 10 * time.Second
)

// ConnectToSomeRunningInstance connects to some running instance.
// It's used for some actions that can be performed via any instance socket,
// no matter if this instance is joined to cluster or not.
// For example, to get known roles or vhsard groups list.
func ConnectToSomeRunningInstance(ctx *context.Ctx) (*connector.Conn, error) {
	instancesConf, err := GetInstancesConf(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get instances configuration: %s", err)
	}

	runningInstancesNames := getRunningInstances(instancesConf, ctx)
	if len(runningInstancesNames) == 0 {
		return nil, fmt.Errorf("No running instances found")
	}

	instanceName := runningInstancesNames[0]
	conn, err := ConnectToInstance(instanceName, ctx)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// ConnectToSomeJoinedInstance connects to some instance joined to cluster.
// It's used for actions with joined cluster, e.g. setting replicaset parameters.
func ConnectToSomeJoinedInstance(ctx *context.Ctx) (*connector.Conn, error) {
	instancesConf, err := GetInstancesConf(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get instances configuration: %s", err)
	}

	joinedInstanceName, err := GetJoinedInstanceName(instancesConf, ctx)
	if err != nil || joinedInstanceName == "" {
		return nil, fmt.Errorf("Failed to find some instance joined to cluster: %s", err)
	}

	conn, err := ConnectToInstance(joinedInstanceName, ctx)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func GetInstancesConf(ctx *context.Ctx) (*InstancesConf, error) {
	var err error

	log.Debugf("Instances configuration file is %s", ctx.Running.RunDir)

	if _, err := os.Stat(ctx.Running.ConfPath); err != nil {
		return nil, fmt.Errorf("Failed to use instances configuration file: %s", err)
	}

	fileContentBytes, err := common.GetFileContentBytes(ctx.Running.ConfPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read instances configuration file: %s", err)
	}

	var allSectionsConf InstancesConf
	if err := yaml.Unmarshal([]byte(fileContentBytes), &allSectionsConf); err != nil {
		return nil, fmt.Errorf("Failed to parse instances configuration file %s: %s", ctx.Running.ConfPath, err)
	}

	instancesConf := make(InstancesConf)

	appInstancePrefix := fmt.Sprintf("%s.", ctx.Project.Name)
	for key, sectionConf := range allSectionsConf {
		if strings.HasPrefix(key, appInstancePrefix) {
			parts := strings.SplitN(key, ".", 2)
			instanceName := parts[1]
			instancesConf[instanceName] = sectionConf
		}
	}

	return &instancesConf, nil
}

// GetJoinedInstanceName returns a name of instancethat is already joined to cluster
// It gets all membership instances and checks if there is some instance that has
// UUID (it means that this instance is joined to cluster).
// The main reason of using membership here is that instances that aren't joined
// can see joined instances only if they are connected to the one membership.
func GetJoinedInstanceName(instancesConf *InstancesConf, ctx *context.Ctx) (string, error) {
	membershipInstances, err := GetMembershipInstances(instancesConf, ctx)

	if err != nil {
		return "", err
	}

	var joinedInstanceName string
	for instanceURI, instance := range *membershipInstances {
		if instance.UUID != "" {
			if instance.Alias == "" {
				return "", fmt.Errorf("Failed to get alias for instance %s", instanceURI)
			}

			joinedInstanceName = instance.Alias
			break
		}
	}

	return joinedInstanceName, nil
}

// GetMembershipInstances returns MembershipInstances for currently configured instances
// First, it connects all instances to membership (probes all running instances by one
// of them). Then, it gets all membership instances members.
func GetMembershipInstances(instancesConf *InstancesConf, ctx *context.Ctx) (*MembershipInstances, error) {
	runningInstancesNames := getRunningInstances(instancesConf, ctx)
	if len(runningInstancesNames) == 0 {
		return nil, fmt.Errorf("No running instances found")
	}

	instanceName := runningInstancesNames[0]
	conn, err := ConnectToInstance(instanceName, ctx)
	if err != nil {
		return nil, err
	}

	log.Debugf("Connect all instances to membership")

	if err := ConnectToMembership(conn, runningInstancesNames, instancesConf); err != nil {
		return nil, fmt.Errorf("Failed to connect instances to membership: %s", err)
	}

	membershipInstances, err := GetMembershipInstancesFromConn(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to get membership instances: %s", err)
	}

	return membershipInstances, nil
}

func getRunningInstances(instancesConf *InstancesConf, ctx *context.Ctx) []string {
	var runningInstancesNames []string
	for instanceName := range *instancesConf {
		process := running.NewInstanceProcess(ctx, instanceName)
		if process.IsRunning() {
			runningInstancesNames = append(runningInstancesNames, instanceName)
		}
	}

	return runningInstancesNames
}

func ConnectToInstance(instanceName string, ctx *context.Ctx) (*connector.Conn, error) {
	consoleSockPath := project.GetInstanceConsoleSock(ctx, instanceName)
	conn, err := connector.Connect(consoleSockPath, connector.Opts{})
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Tarantool instance: %s", err)
	}

	log.Debugf("Connected to %s", consoleSockPath)

	return conn, nil
}

func HealthCheckIsNeeded(conn *connector.Conn) (bool, error) {
	majorCartridgeVersion, err := common.GetMajorCartridgeVersion(conn)
	if err != nil {
		return false, fmt.Errorf("Failed to get Cartridge major version: %s", err)
	}

	return majorCartridgeVersion < 2, nil
}

func ConnectToMembership(conn *connector.Conn, runningInstancesNames []string, instancesConf *InstancesConf) error {
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
