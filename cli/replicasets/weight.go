package replicasets

import (
	"fmt"
	"strconv"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/cluster"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func SetWeight(ctx *context.Ctx, args []string) error {
	if ctx.Replicasets.ReplicasetName == "" {
		return fmt.Errorf("Please, specify replica set name via --replicaset flag")
	}

	if err := FillCtx(ctx); err != nil {
		return err
	}

	if len(args) != 1 {
		return fmt.Errorf("Should be specified one argument - replica set weight")
	}

	weight, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return fmt.Errorf("Failed to parse specified weight. Please, specify valid float")
	}

	conn, err := cluster.ConnectToSomeJoinedInstance(ctx)
	if err != nil {
		return err
	}

	topologyReplicaset, err := getTopologyReplicaset(conn, ctx.Replicasets.ReplicasetName)
	if err != nil {
		return err
	}

	editReplicasetOpts, err := getSetWeightEditReplicasetOpts(weight, topologyReplicaset)
	if err != nil {
		return fmt.Errorf("Failed to get edit_topology options for setting weight: %s", err)
	}

	newTopologyReplicaset, err := editReplicaset(conn, editReplicasetOpts)
	if err != nil {
		return fmt.Errorf("Failed to update roles list: %s", err)
	}

	formattedWeight := strconv.FormatFloat(*newTopologyReplicaset.Weight, 'f', -1, 64)
	log.Infof("Replica set %s weight is set to %s", ctx.Replicasets.ReplicasetName, formattedWeight)

	return nil
}

func getSetWeightEditReplicasetOpts(weight float64, topologyReplicaset *TopologyReplicaset) (*EditReplicasetOpts, error) {
	editReplicasetOpts := EditReplicasetOpts{
		ReplicasetUUID: topologyReplicaset.UUID,
		Weight:         &weight,
	}

	return &editReplicasetOpts, nil
}
