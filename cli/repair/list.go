package repair

import (
	"fmt"
	"sort"
	"strings"

	"github.com/tarantool/cartridge-cli/cli/project"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

const (
	titleWidth = 35

	instancesTitleText   = "Instances"
	replicasetsTitleText = "Replicasets"

	indent = "  "
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

func getIndentedString(n int, format string, a ...interface{}) string {
	return fmt.Sprintf("%s%s", strings.Repeat(indent, n), fmt.Sprintf(format, a...))
}

func getTopologySummary(topologyConf *TopologyConfType, ctx *context.Ctx) ([]common.ResultMessage, error) {
	var resMessages []common.ResultMessage

	failed := false

	// instances
	if instancesSummary, err := getInstancesSummary(topologyConf); err != nil {
		failed = true
		resMessages = append(resMessages, common.GetErrMessage("Failed to get instaces summary: %s", err))
	} else {
		resMessages = append(resMessages, common.GetInfoMessage(instancesSummary))
	}

	// replicasets
	if replicasetsSummary, err := getReplicasetsSummary(topologyConf); err != nil {
		failed = true
		resMessages = append(resMessages, common.GetErrMessage("Failed to get replicasets summary: %s", err))
	} else {
		resMessages = append(resMessages, common.GetInfoMessage(replicasetsSummary))
	}

	resMessages = append(resMessages, common.GetInfoMessage(""))

	if failed {
		return resMessages, fmt.Errorf("Failed to get topology summary")
	}

	return resMessages, nil
}

func getInstancesSummary(topologyConf *TopologyConfType) (string, error) {
	summary := make([]string, 0)

	if len(topologyConf.Instances) == 0 {
		summary = append(summary, common.ColorWarn.Sprintf("No instances found in cluster"))
	} else {
		summary = append(summary, instancesTitle)
	}

	for _, instanceUUID := range topologyConf.GetOrderedInstaceUUIDs() {
		instanceConf, found := topologyConf.Instances[instanceUUID]
		if !found {
			return "", project.InternalError("No instance with UUID %s found", instanceUUID)
		}

		instanceTitle := common.ColorCyan.Sprintf("%s* %s", indent, instanceUUID)

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
			getIndentedString(2, "URI: %s", instanceConf.AdvertiseURI),
			getIndentedString(2, "replicaset: %s", instanceConf.ReplicasetUUID),
		}...)
	}

	return strings.Join(summary, "\n"), nil
}

func getReplicasetsSummary(topologyConf *TopologyConfType) (string, error) {
	summary := make([]string, 0)

	if len(topologyConf.Replicasets) == 0 {
		summary = append(summary, common.ColorWarn.Sprintf("No replicasets found in cluster"))
	} else {
		summary = append(summary, replicasetsTitle)
	}

	for _, replicasetUUID := range topologyConf.GetOrderedReplicasetUUIDs() {
		replicasetConf, found := topologyConf.Replicasets[replicasetUUID]
		if !found {
			return "", project.InternalError("No replicaset with UUID %s found", replicasetUUID)
		}

		replicasetTitle := common.ColorCyan.Sprintf(getIndentedString(1, "* %s", replicasetUUID))

		summary = append(summary, replicasetTitle)

		// alias
		if replicasetConf.Alias != "" {
			summary = append(summary, getIndentedString(2, "alias: %s", replicasetConf.Alias))
		}

		// roles
		if len(replicasetConf.RolesMap) > 0 {
			summary = append(summary, getIndentedString(2, "roles:"))

			// get sorted roles list
			rolesList := make([]string, len(replicasetConf.RolesMap))
			i := 0
			for role := range replicasetConf.RolesMap {
				rolesList[i] = role
				i++
			}
			sort.Strings(rolesList)

			for _, role := range rolesList {
				summary = append(summary, getIndentedString(2, " * %s", role))
			}
		} else {
			summary = append(summary, getIndentedString(2, "No roles"))
		}

		// instances
		summary = append(summary, getIndentedString(2, "instances:"))

		instancesInLeaders := make(map[string]bool)
		for _, leaderUUID := range replicasetConf.Leaders {
			instancesInLeaders[leaderUUID] = true
			summary = append(summary, getIndentedString(2, " * %s", leaderUUID))
		}
		for _, instanceUUID := range replicasetConf.Instances {
			if !instancesInLeaders[instanceUUID] {
				summary = append(summary, getIndentedString(2, " * %s", instanceUUID))
			}
		}
	}

	return strings.Join(summary, "\n"), nil
}
