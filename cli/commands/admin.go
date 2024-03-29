package commands

import (
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tarantool/cartridge-cli/cli/admin"
)

func init() {
	var adminCmd = &cobra.Command{
		Use:   "admin [ADMIN_FUNC_NAME]",
		Short: "Call admin function",
		Long: `Call admin function on application instance
IF --conn flag is specified, CLI connects to instance by specified address.
If --instance flag is specified, then <run-dir>/<app-name>.<instance>.control socket is used.
Otherwise, first available socket from all <run-dir>/<app-name>.*.control is used.`,

		Run: func(cmd *cobra.Command, args []string) {
			err := runAdminCommand(cmd, args)
			if err != nil {
				log.Fatalf(err.Error())
			}
		},
		DisableFlagParsing: true,
	}

	rootCmd.AddCommand(adminCmd)

	// FLAGS are parsed in runAdminCommand
	addAdminFlags(adminCmd.Flags())
}

func addAdminFlags(flagSet *pflag.FlagSet) {
	// add root cmd persistent flags
	flagSet.AddFlagSet(rootCmd.Flags())

	// then, add `cartridge admin` flags
	flagSet.StringVar(&ctx.Project.Name, "name", "", "Application name")
	flagSet.BoolVarP(&ctx.Admin.List, "list", "l", false, "List available admin functions")
	flagSet.BoolVarP(&ctx.Admin.Help, "help", "h", false, "Help for admin function")

	flagSet.StringVar(&ctx.Admin.InstanceName, "instance", "", "Instance to connect to")
	flagSet.StringVarP(&ctx.Admin.ConnString, "conn", "c", "", "Address to connect to")

	flagSet.StringVar(&ctx.Running.RunDir, "run-dir", "", prodRunDirUsage)

	flagSet.SortFlags = false
}

func runAdminCommand(cmd *cobra.Command, args []string) error {
	flagSet := pflag.NewFlagSet("admin", pflag.ContinueOnError)

	addAdminFlags(flagSet)

	// configure flags set
	flagSet.ParseErrorsWhitelist = pflag.ParseErrorsWhitelist{
		UnknownFlags: true,
	}

	if err := flagSet.Parse(args); err != nil {
		return err
	}

	if ctx.Running.RunDir != "" {
		abspath, err := filepath.Abs(ctx.Running.RunDir)
		if err != nil {
			return err
		}
		ctx.Running.RunDir = abspath
	}

	// log level is usually set in rootCmd.PersistentPreRun
	setLogLevel()

	if ctx.Admin.List && !ctx.Admin.Help {
		return admin.Run(admin.List, &ctx, "", nil, nil)
	}

	if len(flagSet.Args()) == 0 {
		// help for `cartridge admin`
		return cmd.Help()
	}

	funcName := strings.Join(flagSet.Args(), ".")

	if ctx.Admin.Help {
		return admin.Run(admin.Help, &ctx, funcName, flagSet, nil)
	}

	return admin.Run(admin.Call, &ctx, funcName, flagSet, args)
}
