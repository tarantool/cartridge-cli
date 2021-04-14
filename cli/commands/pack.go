package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/pack"
)

var (
	packTypeArgs = []string{"tgz", "rpm", "deb", "docker"}
	deps         = []string{}
	depsFile     = ""
)

const defaultPackageDepsFile = "package-deps.txt"

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

	packCmd.Flags().StringSliceVar(&deps, "deps", []string{}, depsUsage)
	packCmd.Flags().StringVar(&depsFile, "deps-file", "", depsFileUsage)
}

func setPackageDeps() error {
	var err error

	if depsFile != "" && len(deps) != 0 {
		return fmt.Errorf("You can't specify --deps and --deps-file flags at the same time")
	}

	if depsFile == "" && len(deps) == 0 {
		defaultPackeDepsFilePath := filepath.Join(ctx.Project.Path, defaultPackageDepsFile)
		if _, err := os.Stat(defaultPackeDepsFilePath); err != nil {
			return fmt.Errorf("Failed to use default package dependencies file: %s", err)
		}

		depsFile = defaultPackeDepsFilePath
	}

	if depsFile != "" {
		if _, err := os.Stat(depsFile); os.IsNotExist(err) {
			return fmt.Errorf("Invalid path to file with dependencies: %s", err)
		}

		content, err := common.GetFileContent(depsFile)
		if err != nil {
			return fmt.Errorf("Failed to get file content: %s", err)
		}

		deps = strings.Split(content, "\n")
	}

	ctx.Pack.Deps, err = common.ParseDependencies(deps)
	if err != nil {
		return fmt.Errorf("Failed to parse dependencies file: %s", err)
	}

	return nil
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

	if ctx.Pack.Type == pack.RpmType || ctx.Pack.Type == pack.DebType {
		if err := setPackageDeps(); err != nil {
			return err
		}
	} else if depsFile != "" || len(deps) != 0 {
		flagName := "deps"
		if depsFile != "" {
			flagName = "deps-file"
		}

		log.Warnf("You specified the --%s flag, but you are not packaging RPM or DEB. "+
			"Flag will be ignored", flagName)
	}

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
