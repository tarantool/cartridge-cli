package commands

import (
	"fmt"
	"os"

	"github.com/apex/log"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/pack"
)

func init() {
	rootCmd.AddCommand(packCmd)

	packCmd.Flags().StringVar(&ctx.Project.Name, "name", "", nameFlagDoc)
	packCmd.Flags().StringVar(&ctx.Pack.Version, "version", "", versionFlagDoc)
	packCmd.Flags().StringVar(&ctx.Pack.Suffix, "suffix", "", suffixFlagDoc)
	packCmd.Flags().StringSliceVar(&ctx.Pack.ImageTags, "tag", []string{}, tagFlagDoc)

	packCmd.Flags().BoolVar(&ctx.Build.InDocker, "use-docker", false, useDockerDoc)
	packCmd.Flags().BoolVar(&ctx.Docker.NoCache, "no-cache", false, noCacheDoc)
	packCmd.Flags().StringVar(&ctx.Build.DockerFrom, "build-from", "", buildFromDoc)
	packCmd.Flags().StringVar(&ctx.Pack.DockerFrom, "from", "", fromDoc)
	packCmd.Flags().StringSliceVar(&ctx.Docker.CacheFrom, "cache-from", []string{}, cacheFromDoc)

	packCmd.Flags().BoolVar(&ctx.Build.SDKLocal, "sdk-local", false, sdkLocalDoc)
	packCmd.Flags().StringVar(&ctx.Build.SDKPath, "sdk-path", "", sdkPathDoc)

	packCmd.Flags().StringVar(&ctx.Pack.UnitTemplatePath, "unit-template", "", unitTemplateFlagDoc)
	packCmd.Flags().StringVar(
		&ctx.Pack.InstUnitTemplatePath, "instantiated-unit-template", "", instUnitTemplateFlagDoc,
	)
	packCmd.Flags().StringVar(
		&ctx.Pack.StatboardUnitTemplatePath, "stateboard-unit-template", "", stateboardUnitTemplateFlagDoc,
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
	ctx.Pack.Type = cmd.Flags().Arg(0)
	ctx.Project.Path = cmd.Flags().Arg(1)
	ctx.Cli.CartridgeTmpDir = os.Getenv(cartridgeTmpDir)

	if err := pack.FillCtx(&ctx); err != nil {
		return err
	}

	if err := checkOptions(&ctx); err != nil {
		return err
	}

	if err := pack.Run(&ctx); err != nil {
		return err
	}

	return nil
}

func checkOptions(ctx *context.Ctx) error {
	if ctx.Pack.Type != pack.RpmType && ctx.Pack.Type != pack.DebType {
		if ctx.Pack.UnitTemplatePath != "" {
			return fmt.Errorf("--unit-template option can be used only with rpm and deb types")
		}

		if ctx.Pack.InstUnitTemplatePath != "" {
			return fmt.Errorf("--instantiated-unit-template option can be used only with rpm and deb types")
		}

		if ctx.Pack.StatboardUnitTemplatePath != "" {
			return fmt.Errorf("--statboard-unit-template option can be used only with rpm and deb types")
		}
	}

	if ctx.Pack.Type != pack.DockerType {
		if len(ctx.Pack.ImageTags) > 0 {
			return fmt.Errorf("--tag option can be used only with docker type")
		}
	}

	if !ctx.Build.InDocker && ctx.Pack.Type != pack.DockerType {
		if len(ctx.Docker.CacheFrom) > 0 {
			return fmt.Errorf("--cache-from option can be used only with --use-docker flag or docker type")
		}

		if ctx.Build.DockerFrom != "" {
			return fmt.Errorf("--build-from option can be used only with --use-docker flag or docker type")
		}

		if ctx.Pack.DockerFrom != "" {
			return fmt.Errorf("--from option can be used only with --use-docker flag or docker type")
		}

		if ctx.Docker.NoCache {
			return fmt.Errorf("--no-cache option can be used only with --use-docker flag or docker type")
		}

		if ctx.Build.SDKLocal {
			return fmt.Errorf("--sdk-local option can be used only with --use-docker flag or docker type")
		}

		if ctx.Build.SDKPath != "" {
			return fmt.Errorf("--sdk-path option can be used only with --use-docker flag or docker type")
		}
	}

	return nil
}

const (
	cartridgeTmpDir = "CARTRIDGE_TEMPDIR"

	nameFlagDoc = `Application name.
The default name comes from the "package"
field in the rockspec file.
`

	versionFlagDoc = `Application version
The default version is determined as the result of
"git describe --tags --long"
`

	suffixFlagDoc = `Result file (or image) name suffix
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

	tagFlagDoc = `Tag(s) of the Docker image that results
from "pack docker"
Used for docker type
`

	fromDoc = `Path to the base Dockerfile of the runtime
image
Defaults to Dockerfile.cartridge
Used for docker type
`

	buildFromDoc = `Path to the base dockerfile fof the build
image
Used on build in docker
Defaults to Dockerfile.build.cartridge
`

	noCacheDoc = `Creates build and runtime images with
"--no-cache" docker flag
`

	cacheFromDoc = `Images to consider as cache sources
for both build and runtime images
See "--cache-from" docker flag
`

	sdkPathDoc = `Path to the SDK to be delivered
in the result artifact
Alternatively, you can pass the path via the
"TARANTOOL_SDK_PATH" environment variable
`

	sdkLocalDoc = `Flag that indicates if SDK from the local
machine should be delivered in the
result artifact
`
)
