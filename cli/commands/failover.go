package commands

import (
	"strings"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/tarantool/cartridge-cli/cli/failover"
)

var (
	failoverModes = []string{"stateful", "eventual", "disabled"}
)

func init() {
	var failoverCmd = &cobra.Command{
		Use:   "failover",
		Short: "Manage application failover",
	}

	rootCmd.AddCommand(failoverCmd)

	var setupCmd = &cobra.Command{
		Use:   "setup",
		Short: "Setup failover with parameters described in a file",

		Args: cobra.ExactValidArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := failover.Setup(&ctx); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	setupCmd.Flags().StringVar(&ctx.Failover.File, "file", "", failoverSetupFileUsage)

	var disableCmd = &cobra.Command{
		Use:   "disable",
		Short: "Disable failover",

		Args: cobra.ExactValidArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := failover.Disable(&ctx); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	var setCmd = &cobra.Command{
		Use:   "set",
		Short: "Setup failover with the specified mod",

		Args: cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			ctx.Failover.Mode = strings.ToLower(cmd.Flags().Arg(0))

			if err := failover.Set(&ctx); err != nil {
				log.Fatalf(err.Error())
			}
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return failoverModes, cobra.ShellCompDirectiveNoFileComp
			}

			return nil, cobra.ShellCompDirectiveDefault
		},
	}

	setCmd.Flags().StringVar(&ctx.Failover.ParamsJSON, "params", "", failoverParamsUsage)
	setCmd.Flags().StringVar(&ctx.Failover.StateProvider, "state-provider", "", stateProviderUsage)
	setCmd.Flags().StringVar(&ctx.Failover.ProviderParamsJSON, "provider-params", "", provdiderParamsUsage)

	failoverSubCommands := []*cobra.Command{
		setupCmd,
		disableCmd,
		setCmd,
	}

	for _, cmd := range failoverSubCommands {
		failoverCmd.AddCommand(cmd)
		configureFlags(cmd)
		addCommonFailoverFlags(cmd)
	}
}
