package failover

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/fatih/structs"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/replicasets"
	"gopkg.in/yaml.v2"
)

const (
	defaultFailoverParamsFile = "failover.yml"
)

func (failoverOpts *FailoverOpts) Manage(ctx *context.Ctx) error {
	conn, err := replicasets.ConnectToSomeRunningInstance(ctx)
	if err != nil {
		return fmt.Errorf("Failed to connect to some instance: %s", err)
	}

	if failoverOpts.StateProvider != nil && *failoverOpts.StateProvider == "stateboard" {
		*failoverOpts.StateProvider = "tarantool"
	}

	req := connector.EvalReq(manageFailoverBody, structs.Map(failoverOpts))
	ok, err := conn.Exec(req)
	if err != nil {
		return fmt.Errorf("Failed to configure failover: %s", err)
	}

	log.Warnf("%s", ok)

	return nil
}

func Setup(ctx *context.Ctx) error {
	var err error

	if err := project.FillCtx(ctx); err != nil {
		return err
	}

	if ctx.Failover.File == "" {
		ctx.Failover.File = defaultFailoverParamsFile
	}

	if ctx.Failover.File, err = filepath.Abs(ctx.Failover.File); err != nil {
		return fmt.Errorf("Failed to get %s failover configuration file absolute path: %s", ctx.Failover.File, err)
	}

	log.Infof("Configure failover described in %s", ctx.Failover.File)

	failoverOpts, err := getFailoverOptsFromFile(ctx)
	if err != nil {
		return fmt.Errorf("Failed to parse %s failover configuration file: %s", ctx.Failover.File, err)
	}

	if err := failoverOpts.Manage(ctx); err != nil {
		return fmt.Errorf("Failed to configure failover: %s", err)
	}

	log.Infof("Failover configured successfully")

	return nil
}

func getFailoverOptsFromFile(ctx *context.Ctx) (*FailoverOpts, error) {
	if _, err := os.Stat(ctx.Failover.File); os.IsNotExist(err) {
		return nil, fmt.Errorf("File %s with failover configurations doesn't exists", ctx.Failover.File)
	} else if err != nil {
		return nil, fmt.Errorf("Failed to process %s file: %s", ctx.Failover.File, err)
	}

	fileContent, err := common.GetFileContentBytes(ctx.Failover.File)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s file: %s", ctx.Failover.File, err)
	}

	var failoverOpts FailoverOpts
	if err := yaml.Unmarshal(fileContent, &failoverOpts); err != nil {
		return nil, fmt.Errorf("Failed to parse failover configurations: %s", err)
	}

	return &failoverOpts, nil
}
