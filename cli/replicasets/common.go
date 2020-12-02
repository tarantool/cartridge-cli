package replicasets

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/running"
	"gopkg.in/yaml.v2"
)

type InstanceConf struct {
	URI string `yaml:"advertise_uri"`
}

type InstancesConf map[string]*InstanceConf

// connectToSomeRunningInstance connects to some running instance.
// It's used for some actions that can be performed via any instance socket,
// no matter if this instance is joined to cluster or not.
// For example, to get known roles or vhsard groups list.
func connectToSomeRunningInstance(ctx *context.Ctx) (net.Conn, error) {
	instancesConf, err := getInstancesConf(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get instances configuration: %s", err)
	}

	runningInstancesNames := getRunningInstances(instancesConf, ctx)
	if len(runningInstancesNames) == 0 {
		return nil, fmt.Errorf("No running instances found")
	}

	instanceName := runningInstancesNames[0]
	conn, err := connectToInstance(instanceName, ctx)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// connectToSomeJoinedInstance connects to some instance joined to cluster.
// It's used for actions with joined cluster, e.g. setting replicaset parameters.
func connectToSomeJoinedInstance(ctx *context.Ctx) (net.Conn, error) {
	instancesConf, err := getInstancesConf(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get instances configuration: %s", err)
	}

	joinedInstanceName, err := getJoinedInstanceName(instancesConf, ctx)
	if err != nil || joinedInstanceName == "" {
		return nil, fmt.Errorf("Failed find some instance joined to cluster")
	}

	conn, err := connectToInstance(joinedInstanceName, ctx)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func getInstancesConf(ctx *context.Ctx) (*InstancesConf, error) {
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

// getJoinedInstanceName returns a name of instancethat is already joined to cluster
// It gets all membership instances and checks if there is some instance that has
// UUID (it means that this instance is joined to cluster).
// The main reason of using membership here is that instances that aren't joined
// can see joined instances only if they are connected to the one membership.
func getJoinedInstanceName(instancesConf *InstancesConf, ctx *context.Ctx) (string, error) {
	membershipInstances, err := getMembershipInstances(instancesConf, ctx)
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

// getMembershipInstances returns MembershipInstances for currently configured instances
// First, it connects all instances to membership (probes all running instances by one
// of them). Then, it gets all membership instances members.
func getMembershipInstances(instancesConf *InstancesConf, ctx *context.Ctx) (*MembershipInstances, error) {
	runningInstancesNames := getRunningInstances(instancesConf, ctx)
	if len(runningInstancesNames) == 0 {
		return nil, fmt.Errorf("No running instances found")
	}

	instanceName := runningInstancesNames[0]
	conn, err := connectToInstance(instanceName, ctx)
	if err != nil {
		return nil, err
	}

	log.Debugf("Connect all replicasets instances to membership")

	if err := connectToMembership(conn, runningInstancesNames, instancesConf); err != nil {
		return nil, fmt.Errorf("Failed to connect instances to membership: %s", err)
	}

	membershipInstances, err := getMembershipInstancesFromConn(conn)
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

func connectToInstance(instanceName string, ctx *context.Ctx) (net.Conn, error) {
	consoleSockPath := project.GetInstanceConsoleSock(ctx, instanceName)
	conn, err := common.ConnectToTarantoolSocket(consoleSockPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Tarantool instance: %s", err)
	}

	log.Debugf("Connected to %s", consoleSockPath)

	return conn, nil
}
