/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
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

const (
	completionsDirName = "completion"

	bash = "bash"
	zsh  = "zsh"
)

type CompletionParams struct {
	DirName  string
	FileName string
}

var (
	completionParams map[string]CompletionParams
)

func init() {
	completionParams = map[string]CompletionParams{
		bash: CompletionParams{
			DirName:  "bash",
			FileName: rootCmd.Name(),
		},
		zsh: CompletionParams{
			DirName:  "zsh",
			FileName: fmt.Sprintf("_%s", rootCmd.Name()),
		},
	}

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
	curDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Failed to get current directory path: %s", err)
	}

	completionsDirPath := filepath.Join(curDir, completionsDirName)
	if err := os.MkdirAll(completionsDirPath, 0755); err != nil {
		return fmt.Errorf("Failed to create completions directory: %s", err)
	}

	for completionType, params := range completionParams {
		completionDirPath := filepath.Join(completionsDirPath, params.DirName)
		if err := os.MkdirAll(completionDirPath, 0755); err != nil {
			return fmt.Errorf("Failed to create %s completions directory: %s", completionType, err)
		}
		completionPath := filepath.Join(completionDirPath, params.FileName)

		switch completionType {
		case bash:
			if err := cmd.Root().GenBashCompletionFile(completionPath); err != nil {
				return fmt.Errorf("failed to generate bash completion: %s", err)
			}
		case zsh:
			if err := cmd.Root().GenZshCompletionFile(completionPath); err != nil {
				return fmt.Errorf("failed to generate zsh completion: %s", err)
			}
		default:
			return fmt.Errorf("Unknown completion type: %s", completionType)
		}

	}

	return nil
}
