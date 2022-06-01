package replicasets

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"time"

	"github.com/apex/log"
	"github.com/avast/retry-go"
	"github.com/tarantool/cartridge-cli/cli/cluster"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
	"gopkg.in/yaml.v2"
)

const (
	vshardRouterRole       = "vshard-router"
	defaultReplicasetsFile = "replicasets.yml"
	instancesFile          = "instances.yml"
)

type ReplicasetConf struct {
	Alias         string   `yaml:"alias,omitempty"`
	InstanceNames []string `yaml:"instances"`
	Roles         []string `yaml:"roles"`

	Weight      *float64 `yaml:"weight,omitempty"`
	AllRW       *bool    `yaml:"all_rw,omitempty"`
	VshardGroup *string  `yaml:"vshard_group,omitempty"`
}

type ReplicasetsConf map[string]*ReplicasetConf

type ReplicasetsList []*ReplicasetConf

func Setup(ctx *context.Ctx, args []string) error {
	var err error

	if err := project.FillCtx(ctx); err != nil {
		return err
	}

	if ctx.Replicasets.File == "" {
		ctx.Replicasets.File = defaultReplicasetsFile
	}
	if ctx.Replicasets.File, err = filepath.Abs(ctx.Replicasets.File); err != nil {
		return fmt.Errorf("Failed to get replicasets configuration file absolute path: %s", err)
	}

	log.Infof("Set up replicasets described in %s", ctx.Replicasets.File)

	replicasetsList, err := getReplicasetsList(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get replicasets configuration: %s", err)
	}

	instancesConf, err := cluster.GetInstancesConf(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get instances configuration: %s", err)
	}

	conn, err := getConnToSetupReplicasets(replicasetsList, instancesConf, ctx)
	if err != nil {
		return err
	}

	topologyReplicasets, err := getTopologyReplicasets(conn)
	if err != nil {
		return fmt.Errorf("Failed to get current topology replicasets: %s", err)
	}

	log.Debugf("Setup replicasets")

	newTopologyReplicasets, err := setupReplicasets(conn, replicasetsList, instancesConf, topologyReplicasets)
	if err != nil {
		return err
	}

	logSetupSummary(topologyReplicasets, newTopologyReplicasets)

	log.Infof("Replicasets are set up successfully")

	if ctx.Replicasets.BootstrapVshard {
		// This step often fails with "no remotes with `vshard-router` role
		// available" error. It happens when `vshard-router` replicaset is created
		// the last (exactly before vshard bootstrapping).
		// Cartridge has `can_bootstrap` function in `cartridge.vshard-utils`, but
		// it can't be used here - in fact it just checks that there are
		// some non-bootstrapped vshard-groups,
		// see https://github.com/tarantool/cartridge/issues/1148.
		// I didn't find a better way to check that cluster is ready for
		// vshard bootstrapping, so I've just added this `magic` retry
		// to prevent confusing error.

		retryOpts := []retry.Option{
			retry.MaxDelay(1 * time.Second),
			retry.Attempts(5),
			retry.LastErrorOnly(true),
		}

		bootstrapVshardFunc := func() error {
			return bootstrapVshard(conn)
		}

		if err := retry.Do(bootstrapVshardFunc, retryOpts...); err != nil {
			return fmt.Errorf("Failed to bootstrap vshard: %s", err)
		}

		log.Infof("Bootstrap vshard task completed successfully, check the cluster status")
	}

	return nil
}

func setupReplicasets(conn *connector.Conn, replicasetsList *ReplicasetsList, instancesConf *cluster.InstancesConf,
	topologyReplicasets *TopologyReplicasets) (*TopologyReplicasets, error) {

	var err error

	newTopologyReplicasets := &TopologyReplicasets{}
	for replicasetUUID, topologyReplicaset := range *topologyReplicasets {
		(*newTopologyReplicasets)[replicasetUUID] = topologyReplicaset
	}

	cartridgeMajorVersion, err := common.GetMajorCartridgeVersion(conn)
	if err != nil {
		return nil, fmt.Errorf("Failed to get Cartridge version: %s", err)
	}

	if cartridgeMajorVersion < 2 && len(*topologyReplicasets) == 0 {
		// create first replicaset with one instance
		// since in old Cartridge bootstrapping cluster from scratch should be
		// performed on a single-server replicaset only

		firstTopologyReplicaset, err := createFirstReplicasetInOldCartridge(conn, replicasetsList, instancesConf)
		if err != nil {
			return nil, err
		}

		(*newTopologyReplicasets)[firstTopologyReplicaset.UUID] = firstTopologyReplicaset
	}

	// create new replicasets and update current
	newTopologyReplicasets, err = createAndUpdateReplicasets(conn, replicasetsList, instancesConf, newTopologyReplicasets)
	if err != nil {
		return nil, err
	}

	// set failover priority
	// This step is performed separately because we need to know instances UUIDs
	// to change failover priority.
	// Generally, we know all instances UUIDs only after creating replicaset or
	// joining new instances to the existing one.
	newTopologyReplicasets, err = setFailoverPriority(conn, replicasetsList, newTopologyReplicasets)
	if err != nil {
		return nil, err
	}

	return newTopologyReplicasets, nil
}

func createAndUpdateReplicasets(conn *connector.Conn, replicasetsList *ReplicasetsList, instancesConf *cluster.InstancesConf,
	topologyReplicasets *TopologyReplicasets) (*TopologyReplicasets, error) {

	editReplicasetsOpts := &EditReplicasetsListOpts{}
	for _, replicasetConf := range *replicasetsList {
		topologyReplicaset := topologyReplicasets.GetByAlias(replicasetConf.Alias)

		if topologyReplicaset == nil {
			editReplicasetOpts, err := getCreateReplicasetEditReplicasetsOpts(replicasetConf, instancesConf)
			if err != nil {
				return nil, fmt.Errorf("Failed to get edit_topology options for creating replicaset: %s", err)
			}
			*editReplicasetsOpts = append(*editReplicasetsOpts, editReplicasetOpts)
		} else {
			editReplicasetOpts, err := getUpdateReplicasetEditReplicasetsOpts(topologyReplicaset, replicasetConf, instancesConf)
			if err != nil {
				return nil, fmt.Errorf("Failed to get edit_topology options for updating replicaset: %s", err)
			}
			*editReplicasetsOpts = append(*editReplicasetsOpts, editReplicasetOpts)
		}
	}

	newTopologyReplicasets, err := editReplicasetsList(conn, editReplicasetsOpts)
	if err != nil {
		return nil, err
	}

	return newTopologyReplicasets, nil
}

func createFirstReplicasetInOldCartridge(conn *connector.Conn, replicasetsList *ReplicasetsList, instancesConf *cluster.InstancesConf) (*TopologyReplicaset, error) {
	firstReplicasetConf := *(*replicasetsList)[0]
	firstReplicasetConf.InstanceNames = firstReplicasetConf.InstanceNames[:1]

	editReplicasetOpts, err := getCreateReplicasetEditReplicasetsOpts(&firstReplicasetConf, instancesConf)
	if err != nil {
		return nil, fmt.Errorf("Failed to get edit_topology options for creating replicaset: %s", err)
	}

	newTopologyReplicaset, err := editReplicaset(conn, editReplicasetOpts)
	if err != nil {
		return nil, err
	}

	if err := waitForClusterIsHealthy(conn); err != nil {
		return nil, fmt.Errorf("Failed to wait for cluster to become healthy: %s", err)
	}

	return newTopologyReplicaset, nil
}

func setFailoverPriority(conn *connector.Conn, replicasetsList *ReplicasetsList, topologyReplicasets *TopologyReplicasets) (*TopologyReplicasets, error) {
	editReplicasetsOpts := EditReplicasetsListOpts{}

	for _, replicasetConf := range *replicasetsList {
		newTopologyReplicaset := topologyReplicasets.GetByAlias(replicasetConf.Alias)

		// set failover priority
		editReplicasetOpts, err := getSetFailoverPriorityEditReplicasetOpts(replicasetConf.InstanceNames, newTopologyReplicaset)
		if err != nil {
			return nil, fmt.Errorf("Failed to get edit_topology options for setting failover priority: %s", err)
		}

		editReplicasetsOpts = append(editReplicasetsOpts, editReplicasetOpts)
	}

	newTopologyReplicasets, err := editReplicasetsList(conn, &editReplicasetsOpts)
	if err != nil {
		return nil, err
	}

	return newTopologyReplicasets, nil
}

func logSetupSummary(topologyReplicasets, newTopologyReplicasets *TopologyReplicasets) {
	for replicasetUUID, newTopologyReplicaset := range *newTopologyReplicasets {
		replicasetID := newTopologyReplicaset.Alias
		if replicasetID == "" {
			replicasetID = newTopologyReplicaset.UUID
		}

		replicasetRes := common.Result{
			ID: replicasetID,
		}

		if oldTopologyReplicaset, found := (*topologyReplicasets)[replicasetUUID]; !found {
			replicasetRes.Status = common.ResStatusCreated
		} else if reflect.DeepEqual(newTopologyReplicaset, oldTopologyReplicaset) {
			replicasetRes.Status = common.ResStatusOk
		} else {
			replicasetRes.Status = common.ResStatusUpdated
		}

		log.Infof(replicasetRes.String())
	}
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

	if len(replicasetsConf) == 0 {
		return nil, fmt.Errorf("No replicasets specified in %s", ctx.Replicasets.File)
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

func getConnToSetupReplicasets(replicasetsList *ReplicasetsList, instancesConf *cluster.InstancesConf, ctx *context.Ctx) (*connector.Conn, error) {
	controlInstanceName, err := cluster.GetJoinedInstanceName(instancesConf, ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to find some instance joined to custer: %s", err)
	}

	if controlInstanceName == "" {
		// get first instance of the first configured replicaset
		if len(*replicasetsList) > 0 {
			controlInstanceName = (*replicasetsList)[0].InstanceNames[0]
		}
	}

	consoleSockPath := project.GetInstanceConsoleSock(ctx, controlInstanceName)
	conn, err := connector.Connect(consoleSockPath, connector.Opts{})
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to Tarantool instance: %s", err)
	}

	log.Debugf("Connected to %s", consoleSockPath)

	return conn, nil
}

func getCreateReplicasetEditReplicasetsOpts(replicasetConf *ReplicasetConf, instancesConf *cluster.InstancesConf) (*EditReplicasetOpts, error) {
	editReplicasetOpts := EditReplicasetOpts{
		ReplicasetAlias: replicasetConf.Alias,
		Roles:           replicasetConf.Roles,
		AllRW:           replicasetConf.AllRW,
		Weight:          replicasetConf.Weight,
		VshardGroup:     replicasetConf.VshardGroup,
	}

	joinInstancesOpts, err := getJoinInstancesOpts(replicasetConf.InstanceNames, instancesConf)
	if err != nil {
		return nil, fmt.Errorf("Failed to get join instances opts: %s", err)
	}
	editReplicasetOpts.JoinInstances = joinInstancesOpts

	return &editReplicasetOpts, nil
}

func getUpdateReplicasetEditReplicasetsOpts(topologyReplicaset *TopologyReplicaset,
	replicasetConf *ReplicasetConf, instancesConf *cluster.InstancesConf) (*EditReplicasetOpts, error) {

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

	newInstancesNames := common.GetStringSlicesDifference(replicasetConf.InstanceNames, topologyReplicasetInstancesAliases)

	joinInstancesOpts, err := getJoinInstancesOpts(newInstancesNames, instancesConf)
	if err != nil {
		return nil, fmt.Errorf("Failed to get join instances opts: %s", err)
	}

	sort.Slice(joinInstancesOpts, func(i, j int) bool {
		return joinInstancesOpts[i].URI < joinInstancesOpts[j].URI
	})

	editReplicasetOpts.JoinInstances = joinInstancesOpts

	return &editReplicasetOpts, nil
}
