package commands

import (
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/failover"
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
			if err := runFailoverCommands(failover.Setup, args); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	addCommonFailoverParamsFlags(setupCmd)

	var disableCmd = &cobra.Command{
		Use:   "disable",
		Short: "Disable failover",

		Args: cobra.ExactValidArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := runFailoverCommands(failover.Disable, args); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	var setEventualCmd = &cobra.Command{
		Use:   "set-eventual",
		Short: "Setup eventual failover",

		Args: cobra.ExactValidArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := runFailoverCommands(failover.RunEventual, args); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	setupCmd.Flags().StringVar(&ctx.Failover.File, "file", "", failoverSetupFileUsage)
	addCommonFailoverParamsFlags(setEventualCmd)

	var setStatefulStateboardCmd = &cobra.Command{
		Use:   "set-stateful-stateboard",
		Short: "Setup stateful stateboard failover",

		Args: cobra.ExactValidArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := runFailoverCommands(failover.RunStatefulStateboard, args); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	setStatefulStateboardCmd.Flags().StringVar(&ctx.Failover.StateboardParams.URI, "uri", "", stateboardURIUsage)
	setStatefulStateboardCmd.Flags().StringVar(&ctx.Failover.StateboardParams.Password, "password", "", stateboardPasswordUsage)

	addCommonFailoverParamsFlags(setStatefulStateboardCmd)

	var setStatefulEtcd2Cmd = &cobra.Command{
		Use:   "set-stateful-etcd2",
		Short: "Setup stateful etcd2 failover",

		Args: cobra.ExactValidArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			if err := runFailoverCommands(failover.RunStatefulEtcd2, args); err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	setStatefulEtcd2Cmd.Flags().StringVar(&ctx.Failover.Etcd2Params.Password, "password", "", etcd2PasswordUsage)
	setStatefulEtcd2Cmd.Flags().StringVar(&ctx.Failover.Etcd2Params.Username, "username", "", etcd2UsernameUsage)
	setStatefulEtcd2Cmd.Flags().StringVar(&ctx.Failover.Etcd2Params.Prefix, "prefix", "", prefixUsage)
	setStatefulEtcd2Cmd.Flags().StringSliceVar(&ctx.Failover.Etcd2Params.Endpoints, "endpoints", nil, endpointsUsage)
	setStatefulEtcd2Cmd.Flags().IntVar(&ctx.Failover.Etcd2Params.LockDelay, "lock-delay", 0, lockDelayUsage)

	addCommonFailoverParamsFlags(setStatefulEtcd2Cmd)

	failoverSubCommands := []*cobra.Command{
		setupCmd,
		disableCmd,
		setEventualCmd,
		setStatefulStateboardCmd,
		setStatefulEtcd2Cmd,
	}

	for _, cmd := range failoverSubCommands {
		failoverCmd.AddCommand(cmd)
		configureFlags(cmd)
		addCommonFailoverFlags(cmd)
	}
}

func runFailoverCommands(failoverFunc func(ctx *context.Ctx, args []string) error, args []string) error {
	if err := failover.FillCtx(&ctx); err != nil {
		return err
	}

	if err := failoverFunc(&ctx, args); err != nil {
		return err
	}

	return nil
}
