package failover

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"gopkg.in/yaml.v2"
)

type FailoverParams struct {
	Mode             string            `yaml:"mode"`
	StateProvider    string            `yaml:"state_provider,omitempty"`
	StateboardParams *StateboardParams `yaml:"stateboard_params,omitempty"`
	Etcd2Params      *Etcd2Params      `yaml:"etcd2_params,omitempty"`

	FailoverTimeout int  `yaml:"failover_timeout,omitempty"`
	FencingEnabled  bool `yaml:"fencing_enabled,omitempty"`
	FencingTimeout  int  `yaml:"fencing_timeout,omitempty"`
	FencingPause    int  `yaml:"fencing_pause,omitempty"`
}

type StateboardParams struct {
	URI      string `yaml:"uri"`
	Password string `yaml:"password"`
}

type Etcd2Params struct {
	Prefix    string   `yaml:"prefix,omitempty"`
	LockDelay int      `yaml:"lock_delay,omitempty"`
	Endpoints []string `yaml:"endpoints,omitempty"`
	Username  string   `yaml:"username,omitempty"`
	Password  string   `yaml:"password,omitempty"`
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

	if err := parseFailoverParams(ctx); err != nil {
		return fmt.Errorf("Failed to parse %s failover configuration file: %s", ctx.Failover.File, err)
	}

	return nil
}

func parseFailoverParams(ctx *context.Ctx) error {
	if _, err := os.Stat(ctx.Failover.File); os.IsNotExist(err) {
		return fmt.Errorf("File %s with failover configurations doesn't exists", ctx.Failover.File)
	} else if err != nil {
		return fmt.Errorf("Failed to process %s file: %s", ctx.Failover.File, err)
	}

	fileContent, err := common.GetFileContentBytes(ctx.Failover.File)
	if err != nil {
		return fmt.Errorf("Failed to read %s file: %s", ctx.Failover.File, err)
	}

	var failoverParams FailoverParams
	if err := yaml.Unmarshal(fileContent, &failoverParams); err != nil {
		return fmt.Errorf("Failed to parse failover configurations: %s", err)
	}

	if failoverParams.Mode != "stateful" && failoverParams.Mode != "eventual" {
		return fmt.Errorf("Failover mode should be `stateful` or `eventual`")
	}

	if failoverParams.Mode == "eventual" {
		if failoverParams.StateProvider != "" {
			return fmt.Errorf(fmt.Sprintf(eventualModeParamsError, "state_provider"))
		}

		if failoverParams.StateboardParams != nil {
			return fmt.Errorf(fmt.Sprintf(eventualModeParamsError, "stateboard_params"))
		}

		if failoverParams.Etcd2Params != nil {
			return fmt.Errorf(fmt.Sprintf(eventualModeParamsError, "etcd2_params"))
		}
	} else {
		if failoverParams.StateProvider != "stateboard" && failoverParams.StateProvider != "etcd2" {
			return fmt.Errorf("Failover `state_provider` should be `stateboard` or `etcd2`")
		}

		if failoverParams.StateProvider == "stateboard" {
			if failoverParams.StateboardParams == nil {
				return fmt.Errorf("You should specify `stateboard_params` when using stateboard provider")
			}

			if failoverParams.Etcd2Params != nil {
				return fmt.Errorf("You shouldn't specify `etcd2_params` when using stateboard provider")
			}

			if failoverParams.StateboardParams.Password == "" || failoverParams.StateboardParams.URI == "" {
				return fmt.Errorf("You should specify `uri` and `password` params when using stateboard provider")
			}
		} else {
			if failoverParams.StateboardParams != nil {
				return fmt.Errorf("You shouldn't specify `stateboard_params` when using etcd2 provider")
			}

			if failoverParams.Etcd2Params != nil && failoverParams.Etcd2Params.LockDelay < 0 {
				return fmt.Errorf(fmt.Sprintf(negativeParamError, "lock_delay"))
			}
		}
	}

	if failoverParams.FailoverTimeout < 0 {
		return fmt.Errorf(fmt.Sprintf(negativeParamError, "FailoverTimeout"))
	}

	if failoverParams.FencingTimeout < 0 {
		return fmt.Errorf(fmt.Sprintf(negativeParamError, "FencingTimeout"))
	}

	if failoverParams.FencingPause < 0 {
		return fmt.Errorf(fmt.Sprintf(negativeParamError, "FencingPause"))
	}

	updateFailoverCtx(&ctx.Failover, &failoverParams)
	log.Warnf("%q aaa %q %q", ctx.Failover, ctx.Failover.StateProvider, ctx.Failover.StateboardParams)

	return nil
}

func updateFailoverCtx(failoverCtx *context.FailoverCtx, failoverParams *FailoverParams) {
	failoverCtx.Mode = failoverParams.Mode
	failoverCtx.StateProvider = failoverParams.StateProvider

	failoverCtx.FailoverTimeout = failoverParams.FailoverTimeout
	failoverCtx.FencingEnabled = failoverParams.FencingEnabled
	failoverCtx.FencingTimeout = failoverParams.FencingTimeout
	failoverCtx.FencingPause = failoverParams.FencingPause

	if failoverParams.Etcd2Params != nil {
		failoverCtx.Etcd2Params = context.Etcd2ParamsCtx{
			Endpoints: failoverParams.Etcd2Params.Endpoints,
			Prefix:    failoverParams.Etcd2Params.Prefix,
			Username:  failoverParams.Etcd2Params.Username,
			Password:  failoverParams.Etcd2Params.Password,
			LockDelay: failoverParams.Etcd2Params.LockDelay,
		}
	}

	if failoverParams.StateboardParams != nil {
		failoverCtx.StateboardParams = context.StateboardParamsCtx{
			URI:      failoverParams.StateboardParams.URI,
			Password: failoverParams.StateboardParams.Password,
		}
	}
}
