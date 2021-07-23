package failover

import (
	"fmt"
)

var (
	eventualModeParamsError = "Please, don't specify --%s flag when using eventual mode"
	defaultStateboardParams = `{"uri": "localhost:4401", "password": "passwd"}`
)

func validateSetFailoverOpts(opts *FailoverOpts) error {
	switch (*opts)["mode"] {
	case "eventual":
		if err := validateEventualMode(opts); err != nil {
			return err
		}
	case "stateful":
		if err := validateStatefulMode(opts); err != nil {
			return err
		}
	case "disabled":
		return nil
	default:
		return fmt.Errorf("Failover mode should be `stateful`, `eventual` or `disabled`")
	}

	return nil
}

func validateEventualMode(opts *FailoverOpts) error {
	if _, found := (*opts)["state_provider"]; found {
		return fmt.Errorf(eventualModeParamsError, "state_provider")
	}

	if _, found := (*opts)["stateboard_params"]; found {
		return fmt.Errorf(eventualModeParamsError, "stateboard_params")
	}

	if _, found := (*opts)["etcd2_params"]; found {
		return fmt.Errorf(eventualModeParamsError, "etcd2_params")
	}

	return nil
}

func validateStatefulMode(opts *FailoverOpts) error {
	if _, found := (*opts)["state_provider"]; !found {
		return fmt.Errorf("Please, specify --state_provider flag when using stateful mode")
	}

	switch (*opts)["state_provider"] {
	case "stateboard":
		if _, found := (*opts)["stateboard_params"]; !found {
			return fmt.Errorf(
				"Please, specify params for stateboard state provider, using --provider-params '%s'",
				defaultStateboardParams,
			)
		}
	case "etcd2":
		return nil // Because all etcd2 parameters are optional
	default:
		return fmt.Errorf("Failover `state_provider` should be `stateboard` or `etcd2`")
	}

	return nil
}
