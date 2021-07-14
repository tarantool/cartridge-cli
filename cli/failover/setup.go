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
	"github.com/tarantool/cartridge-cli/cli/replicasets"
	"gopkg.in/yaml.v2"
)

type FailoverOpts struct {
	Mode             string          `yaml:"mode" structs:"mode"`
	StateProvider    *string         `yaml:"state_provider,omitempty" structs:"state_provider"`
	StateboardParams *ProviderParams `yaml:"stateboard_params,omitempty" structs:"tarantool_params"`
	Etcd2Params      *ProviderParams `yaml:"etcd2_params,omitempty" structs:"etcd2_params"`

	FailoverTimeout *int  `yaml:"failover_timeout,omitempty" json:"failover_timeout" structs:"failover_timeout"`
	FencingEnabled  *bool `yaml:"fencing_enabled,omitempty" json:"fencing_enabled" structs:"fencing_enabled"`
	FencingTimeout  *int  `yaml:"fencing_timeout,omitempty" json:"fencing_timeout" structs:"fencing_timeout"`
	FencingPause    *int  `yaml:"fencing_pause,omitempty" json:"fencing_pause" structs:"fencing_pause"`
}

type ProviderParams map[string]interface{}

func (failoverOpts *FailoverOpts) Manage(ctx *context.Ctx) error {
	conn, err := replicasets.ConnectToSomeRunningInstance(ctx)
	if err != nil {
		return fmt.Errorf("Failed to connect to some instance: %s", err)
	}

	if failoverOpts.StateProvider != nil && *failoverOpts.StateProvider == "stateboard" {
		*failoverOpts.StateProvider = "tarantool"
	}

	result, err := conn.Exec(connector.EvalReq(manageFailoverBody, structs.Map(failoverOpts)))
	if err != nil {
		return fmt.Errorf("Failed to configure failover1: %s", err)
	}

	if len(result) == 2 {
		if funcErr := result[1]; funcErr != nil {
			return fmt.Errorf("Failed to configure failover2: %s", funcErr)
		}
	}

	return nil
}

func Setup(ctx *context.Ctx) error {
	var err error

	if err := FillCtx(ctx); err != nil {
		return err
	}

	if ctx.Failover.File == "" {
		ctx.Failover.File = defaultFailoverFile
	}

	if ctx.Failover.File, err = filepath.Abs(ctx.Failover.File); err != nil {
		return fmt.Errorf("Failed to get %s failover configuration file absolute path: %s", ctx.Failover.File, err)
	}

	log.Infof("Set up failover described in %s", ctx.Failover.File)

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

	var failoverParams FailoverOpts
	if err := yaml.Unmarshal(fileContent, &failoverParams); err != nil {
		return nil, fmt.Errorf("Failed to parse failover configurations: %s", err)
	}

	if err := validateFailoverOpts(&failoverParams); err != nil {
		return nil, err
	}

	return &failoverParams, nil
}
