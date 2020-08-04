package repair

import (
	"fmt"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

const (
	titleWidth = 35

	instancesTitleText   = "Instances"
	replicasetsTitleText = "Replicasets"
)

var (
	instancesTitle   string
	replicasetsTitle string

	expelledMark string
	disabledMark string
)

func init() {
	instancesTitle = common.ColorMagenta.Sprintf(instancesTitleText)
	replicasetsTitle = common.ColorMagenta.Sprintf(replicasetsTitleText)

	expelledMark = common.ColorWarn.Sprintf("expelled")
	disabledMark = common.ColorWarn.Sprintf("disabled")
}

func getTopologySummary(workDir string, ctx *context.Ctx) ([]string, error) {
	summary := make([]string, 0)

	topologyConf, err := getTopologyConf(workDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to get current topology conf: %s", err)
	}

	if ctx.Cli.Verbose {
		summary = append(summary, fmt.Sprintf("Topology config file: %s", topologyConf.Path))
	}

	instancesSummary, err := getInstancesSummary(topologyConf)
	if err != nil {
		summary = append(summary, common.ColorErr.Sprintf("Failed to get instances summary: %s", err))
	}
	summary = append(summary, instancesSummary...)

	summary = append(summary, "")

	replicasetsSummary, err := getReplicasetsSummary(topologyConf)
	if err != nil {
		summary = append(summary, common.ColorErr.Sprintf("Failed to get replicasets summary: %s", err))
	}
	summary = append(summary, replicasetsSummary...)

	return summary, nil
}

func getInstancesSummary(topologyConf *TopologyConfType) ([]string, error) {
	summary := make([]string, 0)

	if len(topologyConf.Instances) == 0 {
		summary = append(summary, common.ColorWarn.Sprintf("No instances found in cluster"))
	} else {
		summary = append(summary, instancesTitle)
	}

	for instanceUUID, instanceConf := range topologyConf.Instances {
		instanceTitle := common.ColorCyan.Sprintf("* %s", instanceUUID)

		if instanceConf.IsExpelled {
			summary = append(summary, fmt.Sprintf("%s %s", instanceTitle, expelledMark))
			continue
		}

		if instanceConf.IsDisabled {
			summary = append(summary, fmt.Sprintf("%s %s", instanceTitle, disabledMark))
		} else {
			summary = append(summary, instanceTitle)
		}

		summary = append(summary, []string{
			fmt.Sprintf("\tURI: %s", instanceConf.AdvertiseURI),
			fmt.Sprintf("\treplicaset: %s", instanceConf.ReplicasetUUID),
		}...)

	}

	return summary, nil
}

func getReplicasetsSummary(topologyConf *TopologyConfType) ([]string, error) {
	summary := make([]string, 0)

	if len(topologyConf.Replicasets) == 0 {
		summary = append(summary, common.ColorWarn.Sprintf("No replicasets found in cluster"))
	} else {
		summary = append(summary, replicasetsTitle)
	}

	for replicasetUUID, ReplicasetConf := range topologyConf.Replicasets {
		replicasetTitle := common.ColorCyan.Sprintf("* %s", replicasetUUID)

		summary = append(summary, replicasetTitle)

		// alias
		if ReplicasetConf.Alias != "" {
			summary = append(summary, fmt.Sprintf("\talias: %s", ReplicasetConf.Alias))
		}

		// roles
		summary = append(summary, "\troles:")
		for _, role := range ReplicasetConf.Roles {
			summary = append(summary, fmt.Sprintf("\t * %s", role))
		}

		// instances
		summary = append(summary, "\tinstances:")

		instancesInLeaders := make(map[string]bool)
		for _, leaderUUID := range ReplicasetConf.Leaders {
			instancesInLeaders[leaderUUID] = true
			summary = append(summary, fmt.Sprintf("\t * %s", leaderUUID))
		}
		for instanceUUID := range ReplicasetConf.Instances {
			if !instancesInLeaders[instanceUUID] {
				summary = append(summary, fmt.Sprintf("\t * %s", instanceUUID))
			}
		}
	}

	return summary, nil
}
