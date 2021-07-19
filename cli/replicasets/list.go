package replicasets

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/cluster"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

const (
	instanceMarker       = "•"
	leaderInstanceMarker = "★"
)

func List(ctx *context.Ctx, args []string) error {
	var err error

	if err := project.FillCtx(ctx); err != nil {
		return err
	}

	conn, err := cluster.ConnectToSomeJoinedInstance(ctx)
	if err != nil {
		return err
	}

	topologyReplicasets, err := getTopologyReplicasets(conn)
	if err != nil {
		return fmt.Errorf("Failed to get current topology replica sets: %s", err)
	}

	replicasetsSummary := getTopologyReplicasetsSummary(topologyReplicasets)

	log.Infof("Current replica sets:\n%s", replicasetsSummary)

	return nil
}

func getTopologyReplicasetsSummary(topologyReplicasets *TopologyReplicasets) string {
	// sort replicasets by aliases
	replicasetsList := make([]*TopologyReplicaset, len(*topologyReplicasets))
	i := 0
	for _, topologyReplicaset := range *topologyReplicasets {
		replicasetsList[i] = topologyReplicaset
		i++
	}

	sort.Slice(replicasetsList, func(i, j int) bool {
		return replicasetsList[i].Alias < replicasetsList[j].Alias
	})

	// get replicasets summaries in sorted aliases order
	replicasetsSummary := make([]string, len(*topologyReplicasets))
	for i, topologyReplicaset := range replicasetsList {
		replicasetsSummary[i] = getTopologyReplicasetSummary(topologyReplicaset)
	}

	return strings.Join(replicasetsSummary, "\n")
}

func getTopologyReplicasetSummary(topologyReplicaset *TopologyReplicaset) string {
	replicasetSummary := []string{}

	// example replicaset summary:
	//
	// • s-1                    default | 123.4 | ALL RW
	// Role: failover-coordinator | vshard-storage | metrics
	// 	★ s1-master localhost:3302        msk
	// 	• s1-replica localhost:3303       spb

	replicasetTitle := fmt.Sprintf(
		"• %s",
		common.ColorHiMagenta.Sprintf(topologyReplicaset.Alias),
	)

	// additionalInfo is vshard group, weight, all rw
	// only specified parameters are shown
	//
	// example:
	// default | 123.4 | ALL RW
	additionalInfo := []string{}
	if topologyReplicaset.VshardGroup != nil {
		additionalInfo = append(additionalInfo, *topologyReplicaset.VshardGroup)
	}
	if topologyReplicaset.Weight != nil {
		formattedWeight := strconv.FormatFloat(*topologyReplicaset.Weight, 'f', -1, 64)
		additionalInfo = append(additionalInfo, formattedWeight)
	}
	if topologyReplicaset.AllRW != nil && *(topologyReplicaset.AllRW) {
		additionalInfo = append(additionalInfo, "ALL RW")
	}

	if len(additionalInfo) > 0 {
		replicasetTitle = fmt.Sprintf(
			"%-30s    %s",
			replicasetTitle,
			common.ColorHiBlue.Sprint(strings.Join(additionalInfo, " | ")),
		)
	}

	// roles
	//
	// example:
	// Role: failover-coordinator | vshard-storage | metrics
	// No roles (if no roles specified)
	var rolesSummary string
	if len(topologyReplicaset.Roles) > 0 {
		rolesSummary = fmt.Sprintf(
			"  Role: %s",
			strings.Join(topologyReplicaset.Roles, " | "),
		)
	} else {
		rolesSummary = "  No roles"
	}

	// instances
	// the leader is marked by start symbol
	// if zone is specified, it's shown too
	//
	// example:
	// ★ s2-master localhost:3304        msk
	// • s2-replica localhost:3305       spb
	instancesSummary := []string{}
	for _, topologyInstance := range topologyReplicaset.Instances {
		instanceMarker := instanceMarker
		if topologyInstance.UUID == topologyReplicaset.LeaderUUID {
			instanceMarker = leaderInstanceMarker
		}

		instanceTitle := fmt.Sprintf(
			"%s %s",
			common.ColorHiCyan.Sprint(topologyInstance.Alias),
			topologyInstance.URI,
		)

		if topologyInstance.Zone != "" {
			instanceTitle = fmt.Sprintf(
				"%-40s %s",
				instanceTitle,
				common.ColorCyan.Sprint(topologyInstance.Zone),
			)
		}

		instanceSummary := fmt.Sprintf(
			"    %s %s",
			instanceMarker,
			instanceTitle,
		)

		instancesSummary = append(instancesSummary, instanceSummary)
	}

	// collect result summary
	replicasetSummary = append(replicasetSummary, replicasetTitle, rolesSummary)
	replicasetSummary = append(replicasetSummary, instancesSummary...)

	return strings.Join(replicasetSummary, "\n")
}
