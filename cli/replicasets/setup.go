package replicasets

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"reflect"

	"github.com/adam-hanna/arrayOperations"
	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
	"gopkg.in/yaml.v2"
)

type ReplicasetConf struct {
	Alias         string
	InstanceNames []string `yaml:"instances"`
	Roles         []string `yaml:"roles"`

	Weight      *float64 `yaml:"weight"`
	AllRW       *bool    `yaml:"all_rw"`
	VshardGroup *string  `yaml:"vshard_group"`
}

type ReplicasetsConf map[string]*ReplicasetConf

type ReplicasetsList []*ReplicasetConf

func Setup(ctx *context.Ctx, args []string) error {
	var err error

	if err := FillCtx(ctx); err != nil {
		return err
	}

	if ctx.Replicasets.File == "" {
		ctx.Replicasets.File = defaultReplicasetsFile
	}
	if ctx.Replicasets.File, err = filepath.Abs(ctx.Replicasets.File); err != nil {
		return fmt.Errorf("Failed to get replicasets configuration file absolute path: %s", err)
	}

	log.Infof("Setting up topology described in %s", ctx.Replicasets.File)

	replicasetsList, err := getReplicasetsList(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get replicasets configuration: %s", err)
	}

	if len(*replicasetsList) == 0 {
		return fmt.Errorf("No replicasets specified in %s", ctx.Replicasets.File)
	}

	instancesConf, err := getInstancesConf(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get instances configuration: %s", err)
	}

	log.Debugf("Find at least one running instance described in %s", ctx.Running.ConfPath)
	controlInstanceName, err := getJoinedInstanceName(instancesConf, ctx)
	if err != nil {
		return fmt.Errorf("Failed find some instance joined to custer")
	}

	if controlInstanceName == "" {
		// get first instance of the first configured replicaset
		if len(*replicasetsList) > 0 {
			controlInstanceName = (*replicasetsList)[0].InstanceNames[0]
		}
	}

	consoleSockPath := project.GetInstanceConsoleSock(ctx, controlInstanceName)
	conn, err := common.ConnectToTarantoolSocket(consoleSockPath)
	if err != nil {
		return fmt.Errorf("Failed to connect to Tarantool instance: %s", err)
	}

	log.Debugf("Connected to %s", consoleSockPath)

	topologyReplicasets, err := getTopologyReplicasets(conn)
	if err != nil {
		return fmt.Errorf("Failed to get current topology replicasets: %s", err)
	}

	var errors []error

	log.Info("Setup replicasets")

	for _, replicasetConf := range *replicasetsList {
		res := common.Result{
			ID: replicasetConf.Alias,
		}

		topologyReplicaset := topologyReplicasets.GetByAlias(replicasetConf.Alias)

		if topologyReplicaset == nil {
			if err := createReplicaset(conn, replicasetConf, instancesConf); err != nil {
				res.Status = common.ResStatusFailed
				res.Error = err
			} else {
				res.Status = common.ResStatusCreated
			}
		} else {
			if changed, err := updateReplicaset(conn, topologyReplicaset, replicasetConf, instancesConf); err != nil {
				res.Status = common.ResStatusFailed
				res.Error = err
			} else if changed {
				res.Status = common.ResStatusUpdated
			} else {
				res.Status = common.ResStatusOk
			}
		}

		if res.Status == common.ResStatusFailed {
			errors = append(errors, res.FormatError())
		}

		log.Infof(res.String())
	}

	if len(errors) > 0 {
		for _, err := range errors {
			log.Errorf("%s", err)
		}
		return fmt.Errorf("Failed to setup replicasets")
	}

	if ctx.Replicasets.BootstrapVshard {
		if err := bootstrapVshard(conn); err != nil {
			return fmt.Errorf("Failed to bootstrap vshard: %s", err)
		}
	}

	log.Infof("Topology is set up successfully")

	return nil
}

func getReplicasetsList(ctx *context.Ctx) (*ReplicasetsList, error) {
	var err error

	if _, err := os.Stat(ctx.Replicasets.File); err != nil {
		return nil, fmt.Errorf("Failed to use replicasets configuration file: %s", err)
	}

	fileContentBytes, err := common.GetFileContentBytes(ctx.Replicasets.File)
	if err != nil {
		return nil, fmt.Errorf("Failed to read replicasets configuration file: %s", err)
	}

	var replicasetsConf ReplicasetsConf
	if err := yaml.Unmarshal([]byte(fileContentBytes), &replicasetsConf); err != nil {
		return nil, fmt.Errorf("Failed to parse replicasets configuration file %s: %s", ctx.Replicasets.File, err)
	}

	replicasetsList := make(ReplicasetsList, len(replicasetsConf))

	i := 0
	for replicasetAlias, replicasetConf := range replicasetsConf {
		replicasetConf.Alias = replicasetAlias

		replicasetsList[i] = replicasetConf
		i++
	}

	return &replicasetsList, nil
}

func createReplicaset(controlConn net.Conn, replicasetConf *ReplicasetConf, instancesConf *InstancesConf) error {
	editReplicasetOpts, err := getCreateReplicasetEditReplicasetsOpts(replicasetConf, instancesConf)
	if err != nil {
		return fmt.Errorf("Failed to get options for edit_topology call: %s", err)
	}

	if _, err := editReplicaset(controlConn, editReplicasetOpts); err != nil {
		return fmt.Errorf("Failed to create replicaset %s: %s", replicasetConf.Alias, err)
	}

	return nil
}

func updateReplicaset(controlConn net.Conn, topologyReplicaset *TopologyReplicaset, replicasetConf *ReplicasetConf, instancesConf *InstancesConf) (bool, error) {
	var err error
	var editReplicasetOpts *EditReplicasetOpts
	var newTopologyReplicaset *TopologyReplicaset

	// join instances and change replicaset params
	editReplicasetOpts, err = getChangeReplicasetEditReplicasetsOpts(topologyReplicaset, replicasetConf, instancesConf)
	if err != nil {
		return false, fmt.Errorf("Failed to get edit_topology options for changing replicaset: %s", err)
	}

	newTopologyReplicaset, err = editReplicaset(controlConn, editReplicasetOpts)
	if err != nil {
		return false, fmt.Errorf("Failed to change replicaset: %s", err)
	}

	// set failover priority
	editReplicasetOpts, err = getSetFailoverPriorityEditReplicasetsOpts(newTopologyReplicaset, replicasetConf.InstanceNames)
	if err != nil {
		return false, fmt.Errorf("Failed to get edit_topology options for setting failover priority: %s", err)
	}

	newTopologyReplicaset, err = editReplicaset(controlConn, editReplicasetOpts)
	if err != nil {
		return false, fmt.Errorf("Failed to set failover priority: %s", err)
	}

	changed := !reflect.DeepEqual(topologyReplicaset, newTopologyReplicaset)

	return changed, nil
}

func getCreateReplicasetEditReplicasetsOpts(replicasetConf *ReplicasetConf, instancesConf *InstancesConf) (*EditReplicasetOpts, error) {
	editReplicasetOpts := EditReplicasetOpts{
		ReplicasetAlias: replicasetConf.Alias,
		Roles:           replicasetConf.Roles,
		AllRW:           replicasetConf.AllRW,
		Weight:          replicasetConf.Weight,
		VshardGroup:     replicasetConf.VshardGroup,
	}

	joinInstancesURIs, err := getInstancesURIs(replicasetConf.InstanceNames, instancesConf)
	if err != nil {
		return nil, fmt.Errorf("Failed to get URIs of a new instances: %s", err)
	}

	editReplicasetOpts.JoinInstancesURIs = *joinInstancesURIs

	return &editReplicasetOpts, nil
}

func getChangeReplicasetEditReplicasetsOpts(topologyReplicaset *TopologyReplicaset,
	replicasetConf *ReplicasetConf, instancesConf *InstancesConf) (*EditReplicasetOpts, error) {

	editReplicasetOpts := EditReplicasetOpts{
		ReplicasetUUID: topologyReplicaset.UUID,
	}

	if replicasetConf.Weight != nil {
		editReplicasetOpts.Weight = replicasetConf.Weight
	}

	if replicasetConf.AllRW != nil {
		editReplicasetOpts.AllRW = replicasetConf.AllRW
	}

	if replicasetConf.VshardGroup != nil {
		editReplicasetOpts.VshardGroup = replicasetConf.VshardGroup
	}

	editReplicasetOpts.Roles = replicasetConf.Roles

	topologyReplicasetInstancesAliases := make([]string, len(topologyReplicaset.Instances))

	for i, instance := range topologyReplicaset.Instances {
		topologyReplicasetInstancesAliases[i] = instance.Alias
	}

	newInstancesNames := arrayOperations.DifferenceString(replicasetConf.InstanceNames, topologyReplicasetInstancesAliases)
	joinInstancesURIs, err := getInstancesURIs(newInstancesNames, instancesConf)
	if err != nil {
		return nil, fmt.Errorf("Failed to get URIs of a new instances: %s", err)
	}
	editReplicasetOpts.JoinInstancesURIs = *joinInstancesURIs

	return &editReplicasetOpts, nil
}
