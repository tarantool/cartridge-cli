package commands

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/src/pack"
	"github.com/tarantool/cartridge-cli/src/project"
)

func init() {
	sdkPathFromEnv := os.Getenv("TARANTOOL_SDK_PATH")

	rootCmd.AddCommand(packCmd)

	packCmd.Flags().StringVar(&projectCtx.Name, "name", "", nameFlagDoc)
	packCmd.Flags().StringVar(&projectCtx.Version, "version", "", versionFlagDoc)
	packCmd.Flags().StringVar(&projectCtx.Suffix, "suffix", "", suffixFlagDoc)
	packCmd.Flags().StringVar(&projectCtx.ImageTag, "tag", "", tagFlagDoc)

	packCmd.Flags().BoolVar(&projectCtx.BuildInDocker, "use-docker", false, useDockerDoc)
	packCmd.Flags().StringVar(&projectCtx.BuildFrom, "build-from", "", buildFromDoc)

	packCmd.Flags().BoolVar(&projectCtx.SDKLocal, "sdk-local", false, sdkLocalDoc)
	packCmd.Flags().StringVar(&projectCtx.SDKPath, "sdk-path", sdkPathFromEnv, sdkPathDoc)

	packCmd.Flags().StringVar(&projectCtx.UnitTemplatePath, "unit-template", "", unitTemplateFlagDoc)
	packCmd.Flags().StringVar(
		&projectCtx.InstUnitTemplatePath, "instantiated-unit-template", "", instUnitTemplateFlagDoc,
	)
	packCmd.Flags().StringVar(
		&projectCtx.StatboardUnitTemplatePath, "stateboard-unit-template", "", stateboardUnitTemplateFlagDoc,
	)
}

var packCmd = &cobra.Command{
	Use:   "pack TYPE [PATH]",
	Short: "Pack application into a distributable bundle",
	Long: `Pack application into a distributable bundle

The supported types are: rpm, tgz, docker, deb`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		err := runPackCommand(cmd, args)
		if err != nil {
			log.Fatalf(err.Error())
		}
	},
}

func runPackCommand(cmd *cobra.Command, args []string) error {
	var err error

	projectCtx.PackType = cmd.Flags().Arg(0)
	projectCtx.Path = cmd.Flags().Arg(1)
	projectCtx.TmpDir = os.Getenv(tmpDirEnv)

	// fill project-specific context
	err = project.FillCtx(&projectCtx)
	if err != nil {
		return err
	}

	// pack project
	err = pack.Run(&projectCtx)
	if err != nil {
		return err
	}

	return nil
}

const (
	tmpDirEnv = "CARTRIDGE_TEMPDIR"

	nameFlagDoc = `Application name.
By default, application name is taken
from the application rockspec.
`

	versionFlagDoc = `Application version
By default, version is discovered by git
`

	suffixFlagDoc = "Result file (or image) name suffix\n"

	tagFlagDoc = `Runtime image tag
Used for docker type
`

	unitTemplateFlagDoc = `Path to the template for systemd
unit file
Used for rpm and deb types
`

	instUnitTemplateFlagDoc = `Path to the template for systemd
instantiated unit file
Used for rpm and deb types
`

	stateboardUnitTemplateFlagDoc = `Path to the template for
stateboard systemd unit file
Used for rpm and deb types
`

	useDockerDoc = `Forces to build the application in Docker`

	buildFromDoc = `Path to the base dockerfile for build image
Used on build in docker
Default to Dockerfile.build.cartridge
`

	sdkLocalDoc = `SDK from the local machine should be
delivered in the result artifact
Used for building in docker with Tarantool Enterprise
`

	sdkPathDoc = `Path to the SDK to be delivered in the result artifact
(env TARANTOOL_SDK_PATH, has lower priority)
Used for building in docker with Tarantool Enterprise
`
)
