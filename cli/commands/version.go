package commands

import (
	"fmt"
	"runtime"
	"strings"

	goVersion "github.com/hashicorp/go-version"
	"github.com/spf13/cobra"
	"github.com/tarantool/cartridge-cli/cli/common"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var (
	gitTag       string
	gitCommit    string
	versionLabel string
)

const (
	unknownVersion = "<unknown>"
	cliName        = "Tarantool Cartridge CLI"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Cartridge CLI",
	Long:  `All software has versions. This is Cartridge CLI's`,
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(buildVersionString(gitTag, gitCommit, versionLabel))
	},
}

func buildVersionString(gitTag, gitCommit, versionLabel string) string {
	var version string

	var versionParts []string
	versionParts = append(versionParts, cliName)

	if gitTag == "" {
		version = unknownVersion
	} else {
		if normalizedVersion, err := goVersion.NewVersion(gitTag); err != nil {
			version = gitTag
		} else {
			version = strings.Join(common.IntsToStrings(normalizedVersion.Segments()), ".")
		}

		if versionLabel != "" {
			version = fmt.Sprintf("%s/%s", version, versionLabel)
		}
	}

	versionStr := fmt.Sprintf("v%s", version)
	versionParts = append(versionParts, versionStr)

	osArchStr := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	versionParts = append(versionParts, osArchStr)

	if gitCommit != "" {
		gitCommitStr := fmt.Sprintf("commit: %s", gitCommit)
		versionParts = append(versionParts, gitCommitStr)
	}

	return strings.Join(versionParts, " ")
}
