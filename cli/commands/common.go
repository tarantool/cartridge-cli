package commands

import (
	"fmt"

	"github.com/spf13/pflag"
)

func setDefaultValue(flags *pflag.FlagSet, name string, value string) error {
	flag := flags.Lookup(name)
	if flag == nil {
		return fmt.Errorf("Failed to find %s flag", name)
	}

	if !flag.Changed {
		flag.Value.Set(value)
	}

	return nil
}
