package repair

import (
	"fmt"
	"strings"
	"time"

	"github.com/tarantool/cartridge-cli/cli/project"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

var (
	repairTimeout = 30 * time.Second
)

type ProcessConfFuncType func(topologyConf *TopologyConfType, ctx *context.Ctx) ([]common.ResultMessage, error)
type PatchConfFuncType func(topologyConf *TopologyConfType, ctx *context.Ctx) error

func List(ctx *context.Ctx) error {
	log.Infof("Get current topology")
	return Run(getTopologySummary, ctx, false)
}

func PatchURI(ctx *context.Ctx) error {
	log.Infof("Set %s advertise URI to %s", ctx.Repair.SetURIInstanceUUID, ctx.Repair.NewURI)
	return Run(patchConfAdvertiseURI, ctx, true)
}

func RemoveInstance(ctx *context.Ctx) error {
	log.Infof("Remove instance with UUID %s", ctx.Repair.RemoveInstanceUUID)
	return Run(patchConfRemoveInstance, ctx, true)
}

func SetLeader(ctx *context.Ctx) error {
	log.Infof("Set %s leader to %s", ctx.Repair.SetLeaderReplicasetUUID, ctx.Repair.SetLeaderInstanceUUID)
	return Run(patchConfSetLeader, ctx, true)
}

func Run(processConfFunc ProcessConfFuncType, ctx *context.Ctx, patchConf bool) error {
	log.Debugf("Data directory is set to: %s", ctx.Running.DataDir)

	// XXX: reload don't work for cartridge < 2.0

	instanceNames, err := getAppInstanceNames(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get application instances working directories: %s", err)
	}

	appConfigs, err := getAppConfigs(instanceNames, ctx)
	if err != nil {
		return fmt.Errorf("Failed to get application cluster-wide configs: %s", err)
	}

	if err := checkConfigsDifferent(&appConfigs, ctx); err != nil {
		return err
	}

	log.Infof("Process application cluster-wide configurations...")
	if err := processConfigs(processConfFunc, &appConfigs, ctx); err != nil {
		return err
	}

	// early-return
	if ctx.Repair.DryRun || !patchConf {
		return nil
	}

	if ctx.Repair.NoReload {
		log.Infof("Write application cluster-wide configurations...")
		log.Warnf("Reloading cluster-wide configurations is skipped")
	} else {
		log.Infof("Write and reload application cluster-wide configurations...")
	}
	if err := writeConfigs(&appConfigs, ctx); err != nil {
		return err
	}

	return nil
}

func checkConfigsDifferent(appConfigs *AppConfigs, ctx *context.Ctx) error {
	if !appConfigs.AreDifferent() {
		return nil
	}

	if !ctx.Repair.Force {
		log.Errorf("Clusterwide config is diverged between instances")
	} else {
		log.Warnf(
			"Clusterwide config is diverged between instances, " +
				"but since --force option is specified, it will be processed anyway. " +
				"Use --verbose option to show config difference between instances",
		)
	}

	if !ctx.Repair.Force || ctx.Cli.Verbose {
		if diffSummary, err := appConfigs.GetDiffs(); err != nil {
			log.Warnf("Failed to get configs difference summary: %s", err)
		} else {
			log.Infof("Configs difference:")
			fmt.Printf("%s\n", diffSummary)
		}
	}

	if !ctx.Repair.Force {
		return fmt.Errorf(
			"Clusterwide config is diverged between instances. " +
				"You can update clusterwide config anyway using --force option",
		)
	}

	return nil
}

func processConfigs(processConfFunc ProcessConfFuncType, appConfigs *AppConfigs, ctx *context.Ctx) error {
	processConfResCh := make(common.ResChan)
	for hash, topologyConf := range appConfigs.confByHash {
		go func(topologyConf *TopologyConfType, hash string, processConfResCh common.ResChan) {
			res := common.Result{
				ID: strings.Join(appConfigs.instancesByHash[hash], ", "),
			}

			messages, err := processConfFunc(topologyConf, ctx)
			if err != nil {
				res.Status = common.ResStatusFailed
				res.Error = err
			} else {
				res.Status = common.ResStatusOk
			}

			res.Messages = messages

			processConfResCh <- res
		}(topologyConf, hash, processConfResCh)
	}

	if err := waitResults(processConfResCh, len(appConfigs.confByHash)); err != nil {
		return fmt.Errorf("Failed to process cluster-wide configurations")
	}

	return nil
}

func writeConfigs(appConfigs *AppConfigs, ctx *context.Ctx) error {
	writeConfResCh := make(common.ResChan)
	for hash, topologyConf := range appConfigs.confByHash {
		for _, instanceName := range appConfigs.instancesByHash[hash] {
			go func(instanceName string, topologyConf *TopologyConfType, writeConfResCh common.ResChan) {
				res := common.Result{
					ID: instanceName,
				}

				topologyConfPath, found := appConfigs.confPathByInstanceID[instanceName]
				if !found {
					res.Status = common.ResStatusFailed
					res.Error = project.InternalError("No config path found for instance %s", instanceName)
				} else {
					// rewrite
					rewriteMessages, err := rewriteConf(topologyConfPath, topologyConf)
					if err != nil {
						res.Status = common.ResStatusFailed
						res.Error = err
					} else {
						res.Status = common.ResStatusOk
					}

					res.Messages = append(res.Messages, rewriteMessages...)

					if !ctx.Repair.NoReload {
						// reload
						reloadMessages, err := reloadConf(topologyConfPath, instanceName, ctx)
						if err != nil {
							res.Status = common.ResStatusFailed
							res.Error = err
						} else {
							res.Status = common.ResStatusOk
						}

						res.Messages = append(res.Messages, reloadMessages...)
					}
				}

				writeConfResCh <- res
			}(instanceName, topologyConf, writeConfResCh)
		}
	}

	if err := waitResults(writeConfResCh, len(appConfigs.confPathByInstanceID)); err != nil {
		return fmt.Errorf("failed to patch some cluster-wide configurations for some instances")
	}

	return nil
}

func waitResults(resCh common.ResChan, resultsN int) error {
	var errors []error

	for i := 0; i < resultsN; i++ {
		select {
		case res := <-resCh:
			log.Info(res.String())

			if res.Status != common.ResStatusOk {
				errors = append(errors, res.FormatError())
			}

			for _, message := range res.Messages {
				switch message.Type {
				case common.ResMessageErr:
					log.Errorf("%s: %s", res.ID, message.Text)
				case common.ResMessageWarn:
					log.Warnf("%s: %s", res.ID, message.Text)
				case common.ResMessageDebug:
					log.Debugf("%s: %s", res.ID, message.Text)
				case common.ResMessageInfo:
					fmt.Println(message.Text)
				default:
					return project.InternalError("Unknown result message type: %d", message.Type)
				}
			}
		case <-time.After(repairTimeout):
			return project.InternalError("Repair timeout %s was reached", repairTimeout)
		}
	}

	if len(errors) > 0 {
		for _, err := range errors {
			log.Errorf("%s", err)
		}
		return fmt.Errorf("Failed to run")
	}

	return nil
}
