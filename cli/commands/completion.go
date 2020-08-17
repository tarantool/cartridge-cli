package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	completionsDirName = "completion"

	bashCompFilePath string
	zshCompFilePath  string

	defaultBashCompFilePath string
	defaultZshCompFilePath  string
)

func init() {
	defaultBashCompFilePath = filepath.Join(completionsDirName, "bash", rootCmd.Name())
	defaultZshCompFilePath = filepath.Join(completionsDirName, "zsh", fmt.Sprintf("_%s", rootCmd.Name()))

	var genCmd = &cobra.Command{
		Hidden: true,
		Use:    "gen",
		Short:  "Generate completion script",
		Args:   cobra.MaximumNArgs(0),
		PreRun: func(cmd *cobra.Command, args []string) {
			cutFlagsDesc(rootCmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := runGenCmd(cmd, args)
			if err != nil {
				log.Fatalf(err.Error())
			}
		},
	}

	rootCmd.AddCommand(genCmd)

	genCmd.Flags().StringVar(&bashCompFilePath, "bash", defaultBashCompFilePath, "Bash completion file path")
	genCmd.Flags().StringVar(&zshCompFilePath, "zsh", defaultZshCompFilePath, "Zsh completion file path")
}

// cutFlagsDesc cuts command usage on first '\n'
// it's needed to make zsh comletion for flags prettier
func cutFlagsDesc(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.VisitAll(func(f *pflag.Flag) {
		f.Usage = strings.SplitN(f.Usage, "\n", 2)[0]
	})

	for _, subCmd := range cmd.Commands() {
		cutFlagsDesc(subCmd)
	}
}

func runGenCmd(cmd *cobra.Command, args []string) error {
	// create directories
	bashCompFileDir := filepath.Dir(bashCompFilePath)
	if err := os.MkdirAll(bashCompFileDir, 0755); err != nil {
		return fmt.Errorf("Failed to create bash completion directory: %s", err)
	}

	zshCompFileDir := filepath.Dir(zshCompFilePath)
	if err := os.MkdirAll(zshCompFileDir, 0755); err != nil {
		return fmt.Errorf("Failed to create zsh completion directory: %s", err)
	}

	// gen completions
	if err := cmd.Root().GenBashCompletionFile(bashCompFilePath); err != nil {
		return fmt.Errorf("failed to generate bash completion: %s", err)
	}

	if err := cmd.Root().GenZshCompletionFile(zshCompFilePath); err != nil {
		return fmt.Errorf("failed to generate zsh completion: %s", err)
	}

	return nil
}
