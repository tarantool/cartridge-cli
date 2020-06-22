package commands

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/pack"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func init() {
	rootCmd.AddCommand(packCmd)

	packCmd.Flags().StringVar(&projectCtx.Name, "name", "", nameFlagDoc)
	packCmd.Flags().StringVar(&projectCtx.Version, "version", "", versionFlagDoc)
	packCmd.Flags().StringVar(&projectCtx.Suffix, "suffix", "", suffixFlagDoc)
	packCmd.Flags().StringSliceVar(&projectCtx.ImageTags, "tag", []string{}, tagFlagDoc)

	packCmd.Flags().BoolVar(&projectCtx.BuildInDocker, "use-docker", false, useDockerDoc)
	packCmd.Flags().BoolVar(&projectCtx.DockerNoCache, "no-cache", false, noCacheDoc)
	packCmd.Flags().StringVar(&projectCtx.BuildFrom, "build-from", "", buildFromDoc)
	packCmd.Flags().StringVar(&projectCtx.From, "from", "", fromDoc)
	packCmd.Flags().StringSliceVar(&projectCtx.DockerCacheFrom, "cache-from", []string{}, cacheFromDoc)

	packCmd.Flags().BoolVar(&projectCtx.SDKLocal, "sdk-local", false, sdkLocalDoc)
	packCmd.Flags().StringVar(&projectCtx.SDKPath, "sdk-path", "", sdkPathDoc)

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
	projectCtx.PackType = cmd.Flags().Arg(0)
	projectCtx.Path = cmd.Flags().Arg(1)
	projectCtx.TmpDir = os.Getenv(tmpDirEnv)

	// fill project-specific context
	if err := project.FillCtx(&projectCtx); err != nil {
		return err
	}

	if err := project.SetSystemRunningPaths(&projectCtx); err != nil {
		return err
	}

	if projectCtx.TarantoolIsEnterprise && (projectCtx.PackType == pack.DockerType || projectCtx.BuildInDocker) {
		if projectCtx.SDKPath == "" {
			sdkPathFromEnv := os.Getenv("TARANTOOL_SDK_PATH")
			projectCtx.SDKPath = sdkPathFromEnv
		}
		if !common.OnlyOneIsTrue(projectCtx.SDKPath != "", projectCtx.SDKLocal) {
			return fmt.Errorf(sdkPathError)
		}
	} else {
		log.Warnf("Specified TARANTOOL_SDK_PATH is ignored")
	}

	if err := checkOptions(&projectCtx); err != nil {
		return err
	}

	// pack project
	if err := pack.Run(&projectCtx); err != nil {
		return err
	}

	return nil
}

func checkOptions(projectCtx *project.ProjectCtx) error {
	if projectCtx.PackType != pack.RpmType && projectCtx.PackType != pack.DebType {
		if projectCtx.UnitTemplatePath != "" {
			return fmt.Errorf("--unit-template option can be used only with rpm and deb types")
		}

		if projectCtx.InstUnitTemplatePath != "" {
			return fmt.Errorf("--instantiated-unit-template option can be used only with rpm and deb types")
		}

		if projectCtx.StatboardUnitTemplatePath != "" {
			return fmt.Errorf("--statboard-unit-template option can be used only with rpm and deb types")
		}
	}

	if projectCtx.PackType != pack.DockerType {
		if len(projectCtx.ImageTags) > 0 {
			return fmt.Errorf("--tag option can be used only with docker type")
		}
	}

	if !projectCtx.BuildInDocker && projectCtx.PackType != pack.DockerType {
		if len(projectCtx.DockerCacheFrom) > 0 {
			return fmt.Errorf("--cache-from option can be used only with --use-docker flag or docker type")
		}

		if projectCtx.BuildFrom != "" {
			return fmt.Errorf("--build-from option can be used only with --use-docker flag or docker type")
		}

		if projectCtx.From != "" {
			return fmt.Errorf("--from option can be used only with --use-docker flag or docker type")
		}

		if projectCtx.DockerNoCache {
			return fmt.Errorf("--no-cache option can be used only with --use-docker flag or docker type")
		}

		if projectCtx.SDKLocal {
			return fmt.Errorf("--sdk-local option can be used only with --use-docker flag or docker type")
		}

		if projectCtx.SDKPath != "" {
			return fmt.Errorf("--sdk-path option can be used only with --use-docker flag or docker type")
		}
	}

	return nil
}

const (
	sdkPathError = `For packing in docker you should specify one of:
	* --sdk-local: to use local SDK
	* --sdk-path: path to SDK
	  (can be passed in environment variable TARANTOOL_SDK_PATH)`

	tmpDirEnv = "CARTRIDGE_TEMPDIR"

	nameFlagDoc = `Application name.
By default, application name is taken
from the application rockspec.
`

	versionFlagDoc = `Application version
By default, version is discovered by git
`

	suffixFlagDoc = "Result file (or image) name suffix\n"

	tagFlagDoc = `Runtime image tag(s)
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

	noCacheDoc = `Create build and runtime images with
--no-cache docker flag
`

	buildFromDoc = `Path to the base dockerfile for build image
Used on build in docker
Default to Dockerfile.build.cartridge
`

	fromDoc = `Path to the base dockerfile for runtime image
Used for docker type
Default to Dockerfile.cartridge
`

	cacheFromDoc = `Images to consider as cache sources
for both build and runtime images
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
