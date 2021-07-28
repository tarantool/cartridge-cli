package failover

import (
	"fmt"
)

var (
	eventualModeParamsError     = "Please, don't specify --%s flag when using eventual mode"
	exampleStateboardParamsJSON = `{"uri": "localhost:4401", "password": "passwd"}`
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
		return fmt.Errorf(eventualModeParamsError, "state-provider")
	}

	return nil
}

func validateStatefulMode(opts *FailoverOpts) error {
	if _, found := (*opts)["state_provider"]; !found {
		return fmt.Errorf("Please, specify --state-provider flag when using stateful mode")
	}

	switch (*opts)["state_provider"] {
	case "stateboard":
		if _, found := (*opts)["stateboard_params"]; !found {
			return fmt.Errorf(
				"Please, specify params for stateboard state provider, using --provider-params '%s'",
				exampleStateboardParamsJSON,
			)
		}
	case "etcd2":
		return nil // Because all etcd2 parameters are optional
	default:
		return fmt.Errorf("--state-provider flag should be `stateboard` or `etcd2`")
	}

	return nil
}
