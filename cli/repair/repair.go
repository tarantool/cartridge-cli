package repair

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/tarantool/cartridge-cli/cli/project"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

var (
	repairTimeout = 30 * time.Second
)

type ProcessConfFuncType func(workDir string, ctx *context.Ctx) ([]string, error)
type PatchConfFuncType func(topologyConf *TopologyConfType, ctx *context.Ctx) error

func PatchURI(ctx *context.Ctx) error {
	log.Infof("Update advertise URI %s -> %s", ctx.Repair.OldURI, ctx.Repair.NewURI)
	return Run(patchConfAdvertiseURI, ctx)
}

func RemoveInstance(ctx *context.Ctx) error {
	log.Infof("Remove instance with UUID %s", ctx.Repair.RemoveInstanceUUID)
	return Run(patchConfRemoveInstance, ctx)
}

func Run(processConfFunc ProcessConfFuncType, ctx *context.Ctx) error {
	appWorkDirNames, err := getAppWorkDirNames(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get application instances working directories: %s", err)
	}

	resCh := make(common.ResChan)

	for _, workDirName := range appWorkDirNames {
		workDirPath := filepath.Join(ctx.Running.DataDir, workDirName)

		go func(workDirPath, workDirName string, resCh common.ResChan) {
			res := common.Result{
				ID: workDirName,
			}

			messages, err := processConfFunc(workDirPath, ctx)
			if err != nil {
				res.Status = common.ResStatusFailed
				res.Error = err
			} else {
				res.Status = common.ResStatusOk
			}

			if ctx.Repair.DryRun || ctx.Cli.Verbose {
				res.Messages = messages
			}

			resCh <- res
		}(workDirPath, workDirName, resCh)
	}

	var errors []error

	for i := 0; i < len(appWorkDirNames); i++ {
		select {
		case res := <-resCh:
			log.Infof(res.String())
			if res.Status != common.ResStatusOk {
				errors = append(errors, res.FormatError())
			}

			if ctx.Repair.DryRun || ctx.Cli.Verbose {
				for _, message := range res.Messages {
					fmt.Println(message)
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
		return fmt.Errorf("Failed to run for some instances")
	}

	return nil
}
