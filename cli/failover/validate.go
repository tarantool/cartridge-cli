package failover

import (
	"fmt"
)

var (
	negativeParamError      = "Parameter %s must be greater than or equal to 0"
	eventualModeParamsError = "You don't have to specify `%s` when using eventual mode"
)

func validateFailoverOpts(opts *FailoverOpts) error {
	switch opts.Mode {
	case "eventual":
		if err := validateEventualMode(opts); err != nil {
			return err
		}
	case "stateful":
		if err := validateStatefulMode(opts); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Failover mode should be `stateful` or `eventual`")
	}

	if err := validateTimeOpts(opts); err != nil {
		return err
	}

	return nil
}

func validateTimeOpts(opts *FailoverOpts) error {
	if opts.FailoverTimeout != nil && *opts.FailoverTimeout < 0 {
		return fmt.Errorf(fmt.Sprintf(negativeParamError, "failover_timeout"))
	}

	if opts.FencingTimeout != nil && *opts.FencingTimeout < 0 {
		return fmt.Errorf(fmt.Sprintf(negativeParamError, "fencing_timeout"))
	}

	if opts.FencingPause != nil && *opts.FencingPause < 0 {
		return fmt.Errorf(fmt.Sprintf(negativeParamError, "fencing_pause"))
	}

	return nil
}

func validateEventualMode(opts *FailoverOpts) error {
	if opts.StateProvider != nil {
		return fmt.Errorf(eventualModeParamsError, "state_provider")
	}

	if opts.StateboardParams != nil {
		return fmt.Errorf(eventualModeParamsError, "stateboard_params")
	}

	if opts.Etcd2Params != nil {
		return fmt.Errorf(eventualModeParamsError, "etcd2_params")
	}

	return nil
}

func validateStatefulMode(opts *FailoverOpts) error {
	if opts.StateProvider == nil {
		return fmt.Errorf("You must specify the `state_provider` when using stateful mode")
	}

	switch *opts.StateProvider {
	case "stateboard":
		if opts.StateboardParams == nil {
			return fmt.Errorf("You should specify `stateboard_params` when using stateboard provider")
		}

		if opts.Etcd2Params != nil {
			return fmt.Errorf("You shouldn't specify `etcd2_params` when using stateboard provider")
		}
	case "etcd2":
		if opts.StateboardParams != nil {
			return fmt.Errorf("You shouldn't specify `stateboard_params` when using etcd2 provider")
		}
	default:
		return fmt.Errorf("Failover `state_provider` should be `stateboard` or `etcd2`")
	}

	return nil
}
