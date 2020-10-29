package pack

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func writeTarantoolTxtFile(file *os.File, content string) {
	if err := ioutil.WriteFile(file.Name(), []byte(content), 0644); err != nil {
		log.Fatalf("Failed to write tarantool.txt file: %s", err)
	}
}

func TestGetTarantoolVersionFromFile(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	var err error
	var ctx = context.Ctx{}

	// create tmp tarantool.txt file
	f, err := ioutil.TempFile("", "tarantool.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	// file is empty
	ctx = context.Ctx{}
	writeTarantoolTxtFile(f, "")
	err = getTarantoolVersionFromFile(f.Name(), &ctx)
	assert.EqualError(
		err,
		"You should specify specify one of TARANTOOL and TARANTOOL_SDK in tarantool.txt file",
	)

	// specified both TARANTOOL and TARANTOOL_SDK
	ctx = context.Ctx{}
	writeTarantoolTxtFile(f, "TARANTOOL=1.2.3-4\nTARANTOOL_SDK=5.6.7-8\n")
	err = getTarantoolVersionFromFile(f.Name(), &ctx)
	assert.EqualError(
		err,
		"You can specify only one of TARANTOOL and TARANTOOL_SDK in tarantool.txt file",
	)

	// bad format
	ctx = context.Ctx{}
	writeTarantoolTxtFile(f, "xxx\n")
	err = getTarantoolVersionFromFile(f.Name(), &ctx)
	assert.EqualError(
		err,
		"tarantool.txt is specified in bad format: could not parse line: xxx",
	)

	// specified TARANTOOL
	tarantoolVersion := "2.3.2-81-g43bcd0b"
	ctx = context.Ctx{}
	writeTarantoolTxtFile(f, fmt.Sprintf("TARANTOOL=%s", tarantoolVersion))
	err = getTarantoolVersionFromFile(f.Name(), &ctx)
	assert.Nil(err)

	assert.Equal(tarantoolVersion, ctx.Tarantool.TarantoolVersion)
	assert.Equal("", ctx.Tarantool.SDKVersion)

	// specified TARANTOOL_SDK
	sdkVersion := "2.3.1-6-g594a358"
	ctx = context.Ctx{}
	writeTarantoolTxtFile(f, fmt.Sprintf("TARANTOOL_SDK=%s", sdkVersion))
	err = getTarantoolVersionFromFile(f.Name(), &ctx)
	assert.Nil(err)

	assert.Equal("", ctx.Tarantool.TarantoolVersion)
	assert.Equal(sdkVersion, ctx.Tarantool.SDKVersion)
}

func TestFillTarantoolCtx(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	var err error
	var ctx = context.Ctx{}

	// create tmp tarantool.txt file
	projectPath, err := ioutil.TempDir("", "project")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(projectPath)

	// get Tarantool from env
	ctxFromEnv := context.Ctx{}
	ctxFromEnv.Project.Path = projectPath
	err = fillTarantoolCtx(&ctxFromEnv)
	assert.Nil(err)

	tarantoolVersionFilePath := filepath.Join(projectPath, "tarantool.txt")
	tarantoolVersionFile, err := os.Create(tarantoolVersionFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// Tarantool version

	tarantoolVersionFromFile := "1.2.3-4-gdeadbee"
	tarantoolVersionSpecified := "3.4.5-6-gdeadbee"

	// --tarantool-version is specified, tarantool.txt exists
	ctx = context.Ctx{}
	ctx.Project.Path = projectPath
	ctx.Tarantool.TarantoolVersion = tarantoolVersionSpecified
	writeTarantoolTxtFile(tarantoolVersionFile, fmt.Sprintf("TARANTOOL=%s", tarantoolVersionFromFile))

	err = fillTarantoolCtx(&ctx)
	assert.Nil(err)
	assert.Equal(tarantoolVersionSpecified, ctx.Tarantool.TarantoolVersion)
	assert.Equal("", ctx.Tarantool.SDKVersion)
	assert.Equal(false, ctx.Tarantool.IsEnterprise)
	assert.Equal(false, ctx.Tarantool.FromEnv)

	// --tarantool-version isn't specified, tarantool.txt exists
	ctx = context.Ctx{}
	ctx.Project.Path = projectPath
	writeTarantoolTxtFile(tarantoolVersionFile, fmt.Sprintf("TARANTOOL=%s", tarantoolVersionFromFile))

	err = fillTarantoolCtx(&ctx)
	assert.Nil(err)
	assert.Equal(tarantoolVersionFromFile, ctx.Tarantool.TarantoolVersion)
	assert.Equal("", ctx.Tarantool.SDKVersion)
	assert.Equal(false, ctx.Tarantool.IsEnterprise)
	assert.Equal(false, ctx.Tarantool.FromEnv)

	// SDK version

	sdkVersionFromFile := "2.3.4-5-gdeadbee"
	sdkVersionSpecified := "4.5.6-7-gdeadbee"

	// --tarantool-sdk is specified, tarantool.txt exists
	ctx = context.Ctx{}
	ctx.Project.Path = projectPath
	ctx.Tarantool.SDKVersion = sdkVersionSpecified
	writeTarantoolTxtFile(tarantoolVersionFile, fmt.Sprintf("TARANTOOL_SDK=%s", sdkVersionFromFile))

	err = fillTarantoolCtx(&ctx)
	assert.Nil(err)
	assert.Equal("", ctx.Tarantool.TarantoolVersion)
	assert.Equal(sdkVersionSpecified, ctx.Tarantool.SDKVersion)
	assert.Equal(true, ctx.Tarantool.IsEnterprise)
	assert.Equal(false, ctx.Tarantool.FromEnv)

	// --sdk-version isn't specified, tarantool.txt exists
	ctx = context.Ctx{}
	ctx.Project.Path = projectPath
	writeTarantoolTxtFile(tarantoolVersionFile, fmt.Sprintf("TARANTOOL_SDK=%s", sdkVersionFromFile))

	err = fillTarantoolCtx(&ctx)
	assert.Nil(err)
	assert.Equal("", ctx.Tarantool.TarantoolVersion)
	assert.Equal(sdkVersionFromFile, ctx.Tarantool.SDKVersion)
	assert.Equal(true, ctx.Tarantool.IsEnterprise)
	assert.Equal(false, ctx.Tarantool.FromEnv)

	// Tarantool from env

	ctx = context.Ctx{}
	ctx.Project.Path = projectPath
	os.RemoveAll(tarantoolVersionFilePath)

	err = fillTarantoolCtx(&ctx)
	assert.Nil(err)
	assert.Equal(ctxFromEnv.Tarantool.TarantoolVersion, ctx.Tarantool.TarantoolVersion)
	assert.Equal(ctxFromEnv.Tarantool.SDKVersion, ctx.Tarantool.SDKVersion)
	assert.Equal(ctxFromEnv.Tarantool.IsEnterprise, ctx.Tarantool.IsEnterprise)
	assert.Equal(true, ctx.Tarantool.FromEnv)
}
