package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
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

func getDuration(durationStr string) (time.Duration, error) {
	if seconds, err := strconv.Atoi(durationStr); err == nil {
		durationStr = fmt.Sprintf("%ds", seconds)
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, err
	}

	if duration < 0 {
		return 0, fmt.Errorf("Negative duration is specified")
	}

	return duration, nil
}

func configureFlags(cmd *cobra.Command) {
	cmd.Flags().SortFlags = false
}

func addNameFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&ctx.Project.Name, "name", "", nameUsage)
}

func addStateboardRunningFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&ctx.Running.WithStateboard, "stateboard", false, stateboardUsage)
	cmd.Flags().BoolVar(&ctx.Running.StateboardOnly, "stateboard-only", false, stateboardOnlyUsage)
}

func addCommonRunningPathsFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&ctx.Running.RunDir, "run-dir", "", runDirUsage)
	cmd.Flags().StringVar(&ctx.Running.ConfPath, "cfg", "", cfgUsage)
}

func addCommonRepairFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&ctx.Project.Name, "name", "", "Application name")
	cmd.Flags().StringVar(&ctx.Running.DataDir, "data-dir", "", repairDataDirUsage)
	cmd.Flags().BoolVar(&ctx.Repair.DryRun, "dry-run", false, dryRunUsage)
}
