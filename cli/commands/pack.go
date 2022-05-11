package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/spf13/cobra"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
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
	packCmd.Flags().StringVar(&ctx.Pack.Filename, "filename", "", filenameUsage)
	packCmd.Flags().StringVar(&ctx.Pack.Suffix, "suffix", "", suffixUsage)
	packCmd.Flags().StringSliceVar(&ctx.Pack.ImageTags, "tag", []string{}, tagUsage)

	packCmd.Flags().BoolVar(&ctx.Build.InDocker, "use-docker", false, useDockerUsage)
	packCmd.Flags().BoolVar(&ctx.Pack.NoCache, "no-cache", false, noCacheUsage)
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
	packCmd.Flags().StringVar(&ctx.Pack.PreInstallScriptFile, "preinst", "", preInstUsage)
	packCmd.Flags().StringVar(&ctx.Pack.PostInstallScriptFile, "postinst", "", postInstUsage)
	packCmd.Flags().StringVar(&ctx.Pack.SystemdUnitParamsPath, "unit-params-file", "", UnitParamsFileUsage)
}

// isExplicitTarantoolDeps returns true if Tarantool was set up by user as a dependency
// with --deps or --deps-file, false otherwise.
func isExplicitTarantoolDeps(deps common.PackDependencies) bool {
	for _, v := range deps {
		if v.Name == "tarantool" {
			return true
		}
	}

	return false
}

// Tarantool dependence is added to rpm and deb packages deps, if it
// wasn't set up explicitly. Dependency conditions is chosen based on
// tarantool version used in cartridge-cli environment. Since development
// builds and entrypoint builds normally are not available in package repos,
// they are not supported as rpm/deb dependency. Minimal required version
// is environment tarantool version, maximum is next tarantool major version.
// Both modern and <= 2.8 version policies are supported.
func addTarantoolDepIfNeeded(ctx *context.Ctx) error {
	var version common.TarantoolVersion
	var minVersion, maxVersion string
	var err error

	if isExplicitTarantoolDeps(ctx.Pack.Deps) {
		return nil
	}

	if ctx.Tarantool.TarantoolIsEnterprise {
		return nil
	}

	if (ctx.Pack.Type != pack.RpmType) && (ctx.Pack.Type != pack.DebType) {
		return nil
	}

	if version, err = common.ParseTarantoolVersion(ctx.Tarantool.TarantoolVersion); err != nil {
		return err
	}

	if version.IsDevelopmentBuild {
		return fmt.Errorf("Development build found. If you want to use Tarantool development build" +
			"as a dependency, set it up explicitly with --deps or --deps-file")
	}

	if version.TagSuffix == "entrypoint" {
		return fmt.Errorf("Entrypoint build found. If you want to use Tarantool entrypoint build" +
			"as a dependency, set it up explicitly with --deps or --deps-file")
	}

	if minVersion, err = common.GetMinimalRequiredVersion(version); err != nil {
		return err
	}
	maxVersion = common.GetNextMajorVersion(version)

	ctx.Pack.Deps.AddTarantool(minVersion, maxVersion)
	return nil
}

func parsePackageDeps(deps []string, depsFile string) (common.PackDependencies, error) {
	var err error

	if depsFile != "" && len(deps) != 0 {
		return nil, fmt.Errorf("You can't specify --deps and --deps-file flags at the same time")
	}

	if depsFile == "" && len(deps) == 0 {
		defaultPackDepsFilePath := filepath.Join(ctx.Project.Path, defaultPackageDepsFile)
		if _, err := os.Stat(defaultPackDepsFilePath); err == nil {
			depsFile = defaultPackDepsFilePath
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("Failed to use default package dependencies file: %s", err)
		}
	}

	if depsFile != "" {
		if _, err := os.Stat(depsFile); os.IsNotExist(err) {
			return nil, fmt.Errorf("Invalid path to file with dependencies: %s", err)
		}

		content, err := common.GetFileContent(depsFile)
		if err != nil {
			return nil, fmt.Errorf("Failed to get file content: %s", err)
		}

		deps = strings.Split(content, "\n")
	}

	parsedDeps, err := common.ParseDependencies(deps)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse dependencies file: %s", err)
	}

	return parsedDeps, nil
}

func fillDependencies(ctx *context.Ctx) error {
	if ctx.Pack.Type == pack.RpmType || ctx.Pack.Type == pack.DebType {
		var err error

		if ctx.Pack.Deps, err = parsePackageDeps(deps, depsFile); err != nil {
			return err
		}

		return addTarantoolDepIfNeeded(ctx)
	}

	if depsFile != "" || len(deps) != 0 {
		flagName := "deps"
		if depsFile != "" {
			flagName = "deps-file"
		}

		log.Warnf("You specified the --%s flag, but you are not packaging RPM or DEB. "+
			"Flag will be ignored", flagName)
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

	if err := pack.Validate(&ctx); err != nil {
		return err
	}

	preOrPostInstScriptIsSet := cmd.Flags().Changed("preinst") || cmd.Flags().Changed("postinst")
	if err := pack.FillCtx(&ctx, preOrPostInstScriptIsSet); err != nil {
		return err
	}

	if err := fillDependencies(&ctx); err != nil {
		return err
	}

	if err := pack.Run(&ctx); err != nil {
		return err
	}

	return nil
}
