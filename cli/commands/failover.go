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

	setStatefulStateboardCmd.Flags().StringVar(&ctx.Failover.StateboardParams.Uri, "uri", "", "TODO")
	setStatefulStateboardCmd.Flags().StringVar(&ctx.Failover.StateboardParams.Password, "password", "", "TODO")

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

	setStatefulEtcd2Cmd.Flags().StringVar(&ctx.Failover.Ectd2Params.Password, "password", "", "TODO")
	setStatefulEtcd2Cmd.Flags().StringVar(&ctx.Failover.Ectd2Params.Username, "username", "", "TODO")
	setStatefulEtcd2Cmd.Flags().StringVar(&ctx.Failover.Ectd2Params.Prefix, "prefix", "", "TODO")
	setStatefulEtcd2Cmd.Flags().StringSliceVar(&ctx.Failover.Ectd2Params.EndPoints, "endpoints", nil, "TODO")
	setStatefulEtcd2Cmd.Flags().IntVar(&ctx.Failover.Ectd2Params.LockDelay, "lock-delay", 0, "TODO")

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
		// addCommonFailoverFlags(cmd)
	}
}

func runFailoverCommands(failoversFunc func(ctx *context.Ctx, args []string) error, args []string) error {
	if err := failover.FillCtx(&ctx); err != nil {
		return err
	}

	/*
		if err := replicasetsFunc(&ctx, args); err != nil {
			return err
		}
	*/

	return nil
}
