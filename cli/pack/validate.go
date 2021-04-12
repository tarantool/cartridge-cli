package pack

import (
	"fmt"

	"github.com/tarantool/cartridge-cli/cli/context"
)

func Validate(ctx *context.Ctx) error {
	if ctx.Pack.Type != RpmType && ctx.Pack.Type != DebType {
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

	if ctx.Pack.Type != DockerType {
		if len(ctx.Pack.ImageTags) > 0 {
			return fmt.Errorf("--tag option can be used only with docker type")
		}
	}

	if !ctx.Build.InDocker && ctx.Pack.Type != DockerType {
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

	if ctx.Pack.Type == RpmType || ctx.Pack.Type == DebType {
		if ctx.Pack.DepsFile != "" && len(ctx.Pack.Deps) != 0 {
			return fmt.Errorf("You can't specify --deps and --deps-file flags at the same time")
		}
	}

	return nil
}
