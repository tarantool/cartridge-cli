package commands

import (
	"fmt"
	"strconv"

	"github.com/apex/log"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/project"
	"github.com/tarantool/cartridge-cli/cli/running"
)

var (
	linesFlagDoc = fmt.Sprintf(`Output the last NUM lines.
Defaults to %d`, defaultLogLines)
)

const (
	defaultLogLines = 15
)

func init() {
	rootCmd.AddCommand(logCmd)

	logCmd.Flags().StringVar(&ctx.Project.Name, "name", "", nameFlagDoc)

	logCmd.Flags().StringVar(&ctx.Running.RunDir, "run-dir", "", runDirFlagDoc)
	logCmd.Flags().StringVar(&ctx.Running.LogDir, "log-dir", "", logDirFlagDoc)
	logCmd.Flags().StringVar(&ctx.Running.ConfPath, "cfg", "", cfgFlagDoc)

	logCmd.Flags().BoolVar(&ctx.Running.WithStateboard, "stateboard", false, stateboardFlagDoc)
	logCmd.Flags().BoolVar(&ctx.Running.StateboardOnly, "stateboard-only", false, stateboardOnlyFlagDoc)

	logCmd.Flags().BoolVarP(&ctx.Running.LogFollow, "follow", "f", false, followFlagDoc)
	logCmd.Flags().IntVarP(&ctx.Running.LogLines, "lines", "n", 0, linesFlagDoc)
}

var logCmd = &cobra.Command{
	Use:   "log [INSTANCE_NAME...]",
	Short: "Get logs of instance(s)",
	Long:  fmt.Sprintf("Get logs of instance(s)n\n%s", runningCommonDoc),
	Run: func(cmd *cobra.Command, args []string) {
		err := runLogCmd(cmd, args)
		if err != nil {
			log.Fatalf(err.Error())
		}
	},
}

func runLogCmd(cmd *cobra.Command, args []string) error {
	fmt.Printf("string(defaultLogLines): %s\n", strconv.Itoa(defaultLogLines))
	if err := setDefaultValue(cmd.Flags(), "lines", strconv.Itoa(defaultLogLines)); err != nil {
		return project.InternalError("Failed to set default lines value: %s", err)
	}

	if err := running.FillCtx(&ctx, args); err != nil {
		return err
	}

	if err := running.Log(&ctx); err != nil {
		return err
	}

	return nil
}

const (
	followFlagDoc = `Output appended data as the file grows
`
)
