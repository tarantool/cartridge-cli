package repair

import (
	"fmt"
	"strings"

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

func getTopologySummary(workDir string, ctx *context.Ctx) ([]common.ResultMessage, error) {
	var resMessages []common.ResultMessage

	topologyConf, err := getTopologyConf(workDir)
	if err != nil {
		return nil, fmt.Errorf("Failed to get current topology conf: %s", err)
	}

	resMessages = append(resMessages, common.GetDebugMessage("Topology config file: %s", topologyConf.Path))

	resMessages = append(resMessages, common.GetInfoMessage(getInstancesSummary(topologyConf)))
	resMessages = append(resMessages, common.GetInfoMessage(getReplicasetsSummary(topologyConf)))

	return resMessages, nil
}

func getInstancesSummary(topologyConf *TopologyConfType) string {
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

	return strings.Join(summary, "\n")
}

func getReplicasetsSummary(topologyConf *TopologyConfType) string {
	summary := make([]string, 0)

	if len(topologyConf.Replicasets) == 0 {
		summary = append(summary, common.ColorWarn.Sprintf("No replicasets found in cluster"))
	} else {
		summary = append(summary, replicasetsTitle)
	}

	for replicasetUUID, replicasetConf := range topologyConf.Replicasets {
		replicasetTitle := common.ColorCyan.Sprintf("* %s", replicasetUUID)

		summary = append(summary, replicasetTitle)

		// alias
		if replicasetConf.Alias != "" {
			summary = append(summary, fmt.Sprintf("\talias: %s", replicasetConf.Alias))
		}

		// roles
		summary = append(summary, "\troles:")
		for _, role := range replicasetConf.Roles {
			summary = append(summary, fmt.Sprintf("\t * %s", role))
		}

		// instances
		summary = append(summary, "\tinstances:")

		instancesInLeaders := make(map[string]bool)
		for _, leaderUUID := range replicasetConf.Leaders {
			instancesInLeaders[leaderUUID] = true
			summary = append(summary, fmt.Sprintf("\t * %s", leaderUUID))
		}
		for _, instanceUUID := range replicasetConf.Instances {
			if !instancesInLeaders[instanceUUID] {
				summary = append(summary, fmt.Sprintf("\t * %s", instanceUUID))
			}
		}
	}

	return strings.Join(summary, "\n")
}
