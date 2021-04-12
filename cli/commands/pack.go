package commands

import (
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/pack"
)

var (
	packTypeArgs = []string{"tgz", "rpm", "deb", "docker"}
)

func init() {
	rootCmd.AddCommand(packCmd)
	configureFlags(packCmd)

	addNameFlag(packCmd)
	addSpecFlag(packCmd)

	packCmd.Flags().StringVar(&ctx.Pack.Version, "version", "", versionUsage)
	packCmd.Flags().StringVar(&ctx.Pack.Suffix, "suffix", "", suffixUsage)
	packCmd.Flags().StringSliceVar(&ctx.Pack.ImageTags, "tag", []string{}, tagUsage)

	packCmd.Flags().BoolVar(&ctx.Build.InDocker, "use-docker", false, useDockerUsage)
	packCmd.Flags().BoolVar(&ctx.Docker.NoCache, "no-cache", false, noCacheUsage)
	packCmd.Flags().StringVar(&ctx.Build.DockerFrom, "build-from", "", buildFromUsage)
	packCmd.Flags().StringVar(&ctx.Pack.DockerFrom, "from", "", fromUsage)
	packCmd.Flags().StringSliceVar(&ctx.Docker.CacheFrom, "cache-from", []string{}, cacheFromUsage)

	packCmd.Flags().BoolVar(&ctx.Build.SDKLocal, "sdk-local", false, sdkLocalUsage)
	packCmd.Flags().StringVar(&ctx.Build.SDKPath, "sdk-path", "", sdkPathUsage)

	packCmd.Flags().StringVar(&ctx.Pack.UnitTemplatePath, "unit-template", "", unitTemplateUsage)
	packCmd.Flags().StringVar(
		&ctx.Pack.InstUnitTemplatePath, "instantiated-unit-template", "", instUnitTemplateUsage,
	)
	packCmd.Flags().StringVar(
		&ctx.Pack.StatboardUnitTemplatePath, "stateboard-unit-template", "", stateboardUnitTemplateUsage,
	)

	packCmd.Flags().StringSliceVar(&ctx.Pack.Deps, "deps", []string{}, depsUsage)
	packCmd.Flags().StringVar(&ctx.Pack.DepsFile, "deps-file", "", depsFileUsage)
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
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return packTypeArgs, cobra.ShellCompDirectiveNoFileComp
		}

		return nil, cobra.ShellCompDirectiveDefault
	},
}

func runPackCommand(cmd *cobra.Command, args []string) error {
	ctx.Pack.Type = strings.ToLower(cmd.Flags().Arg(0))
	ctx.Project.Path = cmd.Flags().Arg(1)
	ctx.Cli.CartridgeTmpDir = os.Getenv(cartridgeTmpDirEnv)

	if err := pack.Validate(&ctx); err != nil {
		return err
	}

	if err := pack.FillCtx(&ctx); err != nil {
		return err
	}

	if err := pack.Run(&ctx); err != nil {
		return err
	}

	return nil
}
