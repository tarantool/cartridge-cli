package repair

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
)

type AppConfigs struct {
	hashes               []string
	instancesByHash      map[string][]string
	confByHash           map[string]*TopologyConfType
	confPathByInstanceID map[string]string
}

func getAppConfigs(instanceNames []string, ctx *context.Ctx) (AppConfigs, error) {
	var appConfigs AppConfigs
	appConfigs.instancesByHash = make(map[string][]string)
	appConfigs.confByHash = make(map[string]*TopologyConfType)
	appConfigs.confPathByInstanceID = make(map[string]string)

	for _, instanceName := range instanceNames {
		workDirPath := project.GetInstanceWorkDir(ctx, instanceName)

		topologyConfPath, err := getTopologyConfPath(workDirPath)
		if err != nil {
			return appConfigs, fmt.Errorf("%s: Failed to get cluster-wide config path: %s", err)
		}

		// if topology config file wasn't found, instance isn't bootstrapped yet,
		// and we just skip it
		if topologyConfPath == "" {
			continue
		}

		appConfigs.confPathByInstanceID[instanceName] = topologyConfPath

		if _, err := os.Stat(topologyConfPath); err != nil {
			return appConfigs, fmt.Errorf("Failed to use topology config: %s", err)
		}

		hash, err := common.FileSHA256Hex(topologyConfPath)
		if err != nil {
			return appConfigs, fmt.Errorf("Failed to get config hash: %s", err)
		}

		appConfigs.instancesByHash[hash] = append(appConfigs.instancesByHash[hash], instanceName)

		if _, found := appConfigs.confByHash[hash]; !found {
			if appConfigs.confByHash[hash], err = getTopologyConf(topologyConfPath); err != nil {
				return appConfigs, fmt.Errorf("Failed to parse topology config %s: %s", topologyConfPath, err)
			}
		}

	}

	if len(appConfigs.confByHash) == 0 {
		return appConfigs, fmt.Errorf("No cluster-wide configs found in %s", ctx.Running.DataDir)
	}

	for _, instanceIDs := range appConfigs.instancesByHash {
		sort.Sort(sort.StringSlice(instanceIDs))
	}

	for hash := range appConfigs.confByHash {
		appConfigs.hashes = append(appConfigs.hashes, hash)
	}
	sort.Sort(sort.StringSlice(appConfigs.hashes))

	return appConfigs, nil
}

func (d *AppConfigs) AreDifferent() bool {
	return len(d.hashes) > 1
}

func (d *AppConfigs) GetDiffs() (string, error) {
	if !d.AreDifferent() {
		return "Configs are equal", nil
	}

	var summaryLines []string

	var hashToCompareWith string
	var configContentToCompareWith []byte

	for _, hash := range d.hashes {
		topologyConf := d.confByHash[hash]
		configContent, err := topologyConf.MarshalContent()
		if err != nil {
			return "", fmt.Errorf("Failed to marshal config content: %s", err)
		}

		if hashToCompareWith == "" {
			hashToCompareWith = hash
			configContentToCompareWith = configContent
			continue
		}

		instancesFrom := strings.Join(d.instancesByHash[hashToCompareWith], ", ")
		instancesTo := strings.Join(d.instancesByHash[hash], ", ")

		linesDiff, err := getDiffLines(
			configContentToCompareWith, configContent,
			instancesFrom, instancesTo,
		)
		if err != nil {
			return "", fmt.Errorf("Failed to get appConfigs difference: %s", err)
		}

		summaryLines = append(summaryLines, linesDiff...)
	}

	return strings.Join(summaryLines, "\n"), nil
}
