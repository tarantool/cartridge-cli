package commands

import (
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/tarantool/cartridge-cli/cli/failover"
)

var (
	statefulJSONParams string
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
	addCommonFailoverParamsFlags(setupCmd)

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
		Short: "Setup failover with parameters",

		Args: cobra.ExactValidArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			setFailoverTimeFlagsIsSet(cmd)
			if err := failover.Set(&ctx, statefulJSONParams); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	setCmd.Flags().StringVar(&ctx.Failover.Mode, "mode", "", modeUsage)
	setCmd.Flags().StringVar(&ctx.Failover.StateProvider, "state-provider", "", stateProviderUsage)
	setCmd.Flags().StringVar(&statefulJSONParams, "stateboard-params", "", "Stateboard parameters in JSON format")
	setCmd.Flags().StringVar(&statefulJSONParams, "etcd2-params", "", "Etcd2 parameters in JSON format")

	addCommonFailoverParamsFlags(setCmd)

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
