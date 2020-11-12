package replicasets

import (
	"fmt"
	"strconv"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func SetWeight(ctx *context.Ctx, args []string) error {
	if ctx.Replicasets.ReplicasetName == "" {
		return fmt.Errorf("Please, specify replicaset name via --replicaset flag")
	}

	if err := FillCtx(ctx); err != nil {
		return err
	}

	if len(args) != 1 {
		return fmt.Errorf("Should be specified one argument - replicaset weight")
	}

	weight, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("Failed to parse specified weight. Please, specify valid float")
	}

	instancesConf, err := getInstancesConf(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get instances configuration: %s", err)
	}

	conn, err := getControlConn(instancesConf, ctx, nil)
	if err != nil {
		return fmt.Errorf("Failed to connect to Tarantool instance: %s", err)
	}

	topologyReplicasets, err := getTopologyReplicasets(conn)
	if err != nil {
		return fmt.Errorf("Failed to get current topology replicasets: %s", err)
	}

	topologyReplicaset := topologyReplicasets.GetByAlias(ctx.Replicasets.ReplicasetName)
	if topologyReplicaset == nil {
		return fmt.Errorf("Replicaset %s isn't found in current topology", ctx.Replicasets.ReplicasetName)
	}

	editReplicasetOpts, err := getSetWeightEditReplicasetsOpts(weight, topologyReplicaset)
	if err != nil {
		return fmt.Errorf("Failed to get edit_topology options for setting weight: %s", err)
	}

	newTopologyReplicaset, err := editReplicaset(conn, editReplicasetOpts)
	if err != nil {
		return fmt.Errorf("Failed to update roles list: %s", err)
	}

	formattedWeight := strconv.FormatFloat(newTopologyReplicaset.Weight, 'f', -1, 64)
	log.Infof("Replicaset %s weight is set to %s", ctx.Replicasets.ReplicasetName, formattedWeight)

	return nil
}

func getSetWeightEditReplicasetsOpts(weight float64, topologyReplicaset *TopologyReplicaset) (*EditReplicasetOpts, error) {
	editReplicasetOpts := EditReplicasetOpts{
		ReplicasetUUID: topologyReplicaset.UUID,
		Weight:         &weight,
	}

	return &editReplicasetOpts, nil
}
