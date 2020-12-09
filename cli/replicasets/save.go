package replicasets

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tarantool/cartridge-cli/cli/project"

	"gopkg.in/yaml.v2"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func Save(ctx *context.Ctx, args []string) error {
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

	log.Infof("Save current replicasets to %s", ctx.Replicasets.File)

	conn, err := connectToSomeJoinedInstance(ctx)
	if err != nil {
		return err
	}

	topologyReplicasets, err := getTopologyReplicasets(conn)
	if err != nil {
		return fmt.Errorf("Failed to get current topology replicasets: %s", err)
	}

	newReplicasetsConf := getReplicasetsConf(topologyReplicasets)
	newConfContent, err := yaml.Marshal(*newReplicasetsConf)
	if err != nil {
		return project.InternalError("Failed to marshal new replicasets conf content: %s", err)
	}

	confFile, err := os.OpenFile(ctx.Replicasets.File, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("Failed to open replicasets config for writing: %s", err)
	}

	if _, err := confFile.Write(newConfContent); err != nil {
		return fmt.Errorf("Failed to write new replicasets config: %s", err)
	}

	return nil
}

func getReplicasetsConf(topologyReplicasets *TopologyReplicasets) *ReplicasetsConf {
	replicasetsConf := &ReplicasetsConf{}

	for _, topologyReplicaset := range *topologyReplicasets {
		replicasetConf := &ReplicasetConf{
			Roles:       topologyReplicaset.Roles,
			AllRW:       topologyReplicaset.AllRW,
			Weight:      topologyReplicaset.Weight,
			VshardGroup: topologyReplicaset.VshardGroup,
		}

		for _, topologyInstance := range topologyReplicaset.Instances {
			replicasetConf.InstanceNames = append(replicasetConf.InstanceNames, topologyInstance.Alias)
		}

		(*replicasetsConf)[topologyReplicaset.Alias] = replicasetConf
	}

	return replicasetsConf
}
