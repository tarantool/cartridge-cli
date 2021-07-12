package failover

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/connector"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/replicasets"
	"gopkg.in/yaml.v2"
)

type FailoverOpts struct {
	Mode             string          `yaml:"mode"`
	StateProvider    *string         `yaml:"state_provider,omitempty"`
	StateboardParams *StateboardOpts `yaml:"stateboard_params,omitempty"`
	Etcd2Params      *Etcd2Opts      `yaml:"etcd2_params,omitempty"`

	FailoverTimeout *int  `yaml:"failover_timeout,omitempty"`
	FencingEnabled  *bool `yaml:"fencing_enabled,omitempty"`
	FencingTimeout  *int  `yaml:"fencing_timeout,omitempty"`
	FencingPause    *int  `yaml:"fencing_pause,omitempty"`
}

type StateboardOpts struct {
	URI      string `yaml:"uri"`
	Password string `yaml:"password"`
}

type Etcd2Opts struct {
	Prefix    *string  `yaml:"prefix,omitempty"`
	LockDelay *int     `yaml:"lock_delay,omitempty"`
	Endpoints []string `yaml:"endpoints,omitempty"`
	Username  *string  `yaml:"username,omitempty"`
	Password  *string  `yaml:"password,omitempty"`
}

var (
	negativeParamError      = "Parameter %s must be greater than or equal to 0"
	eventualModeParamsError = "You don't have to specify `%s` when using eventual mode"
)

func Setup(ctx *context.Ctx, args []string) error {
	var err error

	if ctx.Failover.File == "" {
		ctx.Failover.File = defaultFailoverFile
	}

	if ctx.Failover.File, err = filepath.Abs(ctx.Failover.File); err != nil {
		return fmt.Errorf("Failed to get %s failover configuration file absolute path: %s", ctx.Failover.File, err)
	}

	log.Infof("Set up failover described in %s", ctx.Failover.File)

	failoverOpts, err := getFailoverOpts(ctx)
	if err != nil {
		return fmt.Errorf("Failed to parse %s failover configuration file: %s", ctx.Failover.File, err)
	}

	if failoverOpts.Mode == "disable" {
		if err := disableFailover(ctx); err != nil {
			return fmt.Errorf("Failed to disable failover: %s", err)
		}
	} else {
		if err := setupFailover(ctx, failoverOpts); err != nil {
			return fmt.Errorf("Failed to configure failover: %s", err)
		}
	}

	log.Infof("Failover configured successfully")

	return nil
}

func setupFailover(ctx *context.Ctx, failoverOpts *FailoverOpts) error {
	conn, err := replicasets.ConnectToSomeRunningInstance(ctx)
	if err != nil {
		return fmt.Errorf("Failed to connect to some instance: %s", err)
	}

	req := connector.EvalReq(setupFailoverBody, common.StructToMapWithoutNils(failoverOpts))
	if _, err := conn.Exec(req); err != nil {
		return fmt.Errorf("Failed to configure failover: %s", err)
	}

	return nil
}

func disableFailover(ctx *context.Ctx) error {
	return nil
}

func getFailoverOpts(ctx *context.Ctx) (*FailoverOpts, error) {
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

	if failoverParams.Mode != "stateful" && failoverParams.Mode != "eventual" && failoverParams.Mode != "disabled" {
		return nil, fmt.Errorf("Failover mode should be `stateful`, `eventual` or `disabled`")
	}

	if failoverParams.Mode == "disabled" {
		return &failoverParams, nil
	}

	provider := failoverParams.StateProvider
	if failoverParams.Mode == "eventual" {
		if provider != nil {
			return nil, fmt.Errorf(fmt.Sprintf(eventualModeParamsError, "state_provider"))
		}

		if failoverParams.StateboardParams != nil {
			return nil, fmt.Errorf(fmt.Sprintf(eventualModeParamsError, "stateboard_params"))
		}

		if failoverParams.Etcd2Params != nil {
			return nil, fmt.Errorf(fmt.Sprintf(eventualModeParamsError, "etcd2_params"))
		}
	} else {
		if provider == nil || (*provider != "stateboard" && *provider != "etcd2") {
			return nil, fmt.Errorf("Failover `state_provider` should be `stateboard` or `etcd2`")
		}

		if *provider == "stateboard" {
			if failoverParams.StateboardParams == nil {
				return nil, fmt.Errorf("You should specify `stateboard_params` when using stateboard provider")
			}

			if failoverParams.Etcd2Params != nil {
				return nil, fmt.Errorf("You shouldn't specify `etcd2_params` when using stateboard provider")
			}

			if failoverParams.StateboardParams.Password == "" || failoverParams.StateboardParams.URI == "" {
				return nil, fmt.Errorf("You should specify `uri` and `password` params when using stateboard provider")
			}
		} else {
			if failoverParams.Etcd2Params == nil {
				return nil, fmt.Errorf("You should specify `etcd2_params` when using stateboard provider")
			}

			if failoverParams.StateboardParams != nil {
				return nil, fmt.Errorf("You shouldn't specify `stateboard_params` when using etcd2 provider")
			}

			if failoverParams.Etcd2Params != nil && *failoverParams.Etcd2Params.LockDelay < 0 {
				return nil, fmt.Errorf(fmt.Sprintf(negativeParamError, "lock_delay"))
			}
		}
	}

	if failoverParams.FailoverTimeout != nil && *failoverParams.FailoverTimeout < 0 {
		return nil, fmt.Errorf(fmt.Sprintf(negativeParamError, "failover_timeout"))
	}

	if failoverParams.FencingTimeout != nil && *failoverParams.FencingTimeout < 0 {
		return nil, fmt.Errorf(fmt.Sprintf(negativeParamError, "fencing_timeout"))
	}

	if failoverParams.FencingPause != nil && *failoverParams.FencingPause < 0 {
		return nil, fmt.Errorf(fmt.Sprintf(negativeParamError, "fencing_pause"))
	}

	return &failoverParams, nil
}
