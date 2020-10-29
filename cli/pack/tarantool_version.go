package pack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tarantool/cartridge-cli/cli/project"

	"github.com/apex/log"
	"github.com/robfig/config"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

const (
	tarantoolVersionFileName = "tarantool.txt"
	tarantoolVersionOptName  = "TARANTOOL"
	sdkVersionOptName        = "TARANTOOL_SDK"
)

func fillTarantoolCtx(ctx *context.Ctx) error {
	if err := getTarantoolVersion(ctx); err != nil {
		return fmt.Errorf("Failed to get Tarantool version: %s", err)
	}

	if ctx.Tarantool.SDKVersion != "" {
		ctx.Tarantool.IsEnterprise = true
		log.Debugf("SDK %s is used for result artifact", ctx.Tarantool.SDKVersion)
	} else if ctx.Tarantool.TarantoolVersion != "" {
		log.Debugf("Check that Tarantool version is correct")

		if err := common.CheckTarantoolVersion(ctx.Tarantool.TarantoolVersion); err != nil {
			return fmt.Errorf("Detected Tarantool version can't be used: %s", err)
		}

		log.Debugf("Tarantool %s is used for result artifact", ctx.Tarantool.TarantoolVersion)
	} else {
		return project.InternalError(
			"One of ctx.Tarantool.SDKVersion and ctx.Tarantool.TarantoolVersion should be set",
		)
	}

	if !ctx.Tarantool.FromEnv && !ctx.Build.InDocker {
		// warn if application is packed with different version of Tarantool
	}

	return nil
}

func getTarantoolVersion(ctx *context.Ctx) error {
	if ctx.Tarantool.TarantoolVersion != "" {
		// Tarantool version is specified by user
		log.Debugf("Specified Tarantool version")
		return nil
	} else if ctx.Tarantool.SDKVersion != "" {
		// SDK version is specified by user
		log.Debugf("Specified SDK version")
		return nil
	}

	// Tarantool or SDK version is specified in tarantool.txt
	tarantoolVersionFilePath := filepath.Join(ctx.Project.Path, tarantoolVersionFileName)
	if _, err := os.Stat(tarantoolVersionFilePath); err == nil {
		log.Debugf("Found %s file, get Tarantool version from it...", tarantoolVersionFileName)
		if err := getTarantoolVersionFromFile(tarantoolVersionFilePath, ctx); err != nil {
			return fmt.Errorf("Failed to get Tarantool version from %s: %s", tarantoolVersionFileName, err)
		}

		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Failed to use %s file: %s", tarantoolVersionFileName, err)
	}

	log.Debugf("Detect Tarantool version from env...")
	if err := getTarantoolVersionFromEnv(ctx); err != nil {
		return fmt.Errorf(
			"Failed to get Tarantool version from env: %s. "+
				"Please, specify --tarantool-version or --sdk-vesion flag, "+
				"or add %s file to your project", tarantoolVersionFileName,
			err,
		)
	}

	return nil
}

func getTarantoolVersionFromFile(tarantoolVersionFilePath string, ctx *context.Ctx) error {
	c, err := config.ReadDefault(tarantoolVersionFilePath)
	if err != nil {
		return fmt.Errorf("%s is specified in bad format: %s", tarantoolVersionFileName, err)
	}

	var tarantoolVersion string
	var sdkVersion string

	tarantoolVersion, _ = c.RawStringDefault(tarantoolVersionOptName)
	sdkVersion, _ = c.RawStringDefault(sdkVersionOptName)

	if tarantoolVersion != "" && sdkVersion != "" {
		return fmt.Errorf(
			"You can specify only one of %s and %s in %s file",
			tarantoolVersionOptName, sdkVersionOptName, tarantoolVersionFileName,
		)
	}

	if tarantoolVersion == "" && sdkVersion == "" {
		return fmt.Errorf(
			"You should specify specify one of %s and %s in %s file",
			tarantoolVersionOptName, sdkVersionOptName, tarantoolVersionFileName,
		)
	}

	ctx.Tarantool.TarantoolVersion = tarantoolVersion
	ctx.Tarantool.SDKVersion = sdkVersion

	return nil
}

func getTarantoolVersionFromEnv(ctx *context.Ctx) error {
	var err error

	ctx.Tarantool.TarantoolDir, err = common.GetTarantoolDir()
	if err != nil {
		return fmt.Errorf("Failed to find Tarantool executable: %s", err)
	}

	tarantoolIsEnterprise, err := common.TarantoolIsEnterprise(ctx.Tarantool.TarantoolDir)
	if err != nil {
		return fmt.Errorf("Failed to check Tarantool version: %s", err)
	}

	if !tarantoolIsEnterprise {
		ctx.Tarantool.TarantoolVersion, err = common.GetTarantoolVersion(ctx.Tarantool.TarantoolDir)
		if err != nil {
			return fmt.Errorf("Failed to get Tarantool version: %s", err)
		}
	} else {
		ctx.Tarantool.SDKVersion, err = common.GetSDKVersion(ctx.Tarantool.TarantoolDir)
		if err != nil {
			return fmt.Errorf("Failed to get SDK version: %s", err)
		}
	}

	ctx.Tarantool.FromEnv = true

	return nil
}
