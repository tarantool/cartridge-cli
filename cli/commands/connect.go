package commands

import (
	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/tarantool/cartridge-cli/cli/connect"
)

func init() {
	var enterCmd = &cobra.Command{
		Use:   "enter INSTANCE_NAME",
		Short: "Enter to application instance console",
		Run: func(cmd *cobra.Command, args []string) {
			if err := connect.Enter(&ctx, args); err != nil {
				log.Fatalf(err.Error())
			}
		},
		ValidArgsFunction: ShellCompRunningInstances,
		Args:              cobra.MaximumNArgs(1),
	}

	rootCmd.AddCommand(enterCmd)

	// FLAGS
	configureFlags(enterCmd)

	// application name flag
	addNameFlag(enterCmd)
	// run-dir flag
	enterCmd.Flags().StringVar(&ctx.Running.RunDir, "run-dir", "", runDirUsage)

	var connectCmd = &cobra.Command{
		Use:   "connect URI",
		Short: "Connect to specified URI",
		Run: func(cmd *cobra.Command, args []string) {
			if err := connect.Connect(&ctx, args); err != nil {
				log.Fatalf(err.Error())
			}
		},
		Args: cobra.MaximumNArgs(1),
	}

	rootCmd.AddCommand(connectCmd)

	// FLAGS
	configureFlags(connectCmd)

	// username flag
	connectCmd.Flags().StringVarP(&ctx.Connect.Username, "username", "u", "", connectUsernameUsage)
	// password flag
	connectCmd.Flags().StringVarP(&ctx.Connect.Password, "password", "p", "", connectPasswordUsage)
}
