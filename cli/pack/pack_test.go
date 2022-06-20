package pack

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func TestGenerateVersionFileNameEE(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	var ctx context.Ctx

	ctx.Project.Name = "myapp"
	ctx.Pack.Version = "1.2.3.4"
	ctx.Tarantool.TarantoolIsEnterprise = true
	ctx.Build.InDocker = true

	dir, err := ioutil.TempDir("", "__temporary_sdk")
	assert.Equal(err, nil)
	defer os.RemoveAll(dir)

	ctx.Build.SDKPath = dir
	versionFileLines := []string{
		"TARANTOOL=2.8.1-0-ge2a1ec0c2-r409",
		"TARANTOOL_SDK=2.8.1-0-ge2a1ec0c2-r409",
	}

	tmpVersion := filepath.Join(dir, "VERSION")
	err = ioutil.WriteFile(tmpVersion, []byte(strings.Join(versionFileLines, "\n")), 0666)
	assert.Nil(err)

	err = generateVersionFile("", &ctx)
	defer os.Remove("VERSION")
	assert.Nil(err)

	content, err := ioutil.ReadFile("VERSION")
	assert.Nil(err)

	verStr := fmt.Sprintf("%s=%s", ctx.Project.Name, ctx.Pack.VersionWithSuffix)
	expFileLines := append([]string{verStr}, versionFileLines...)
	assert.Equal(expFileLines, strings.Split(string(content), "\n")[:3])
}

func writeTarantoolTxtFile(file *os.File, content string, assert *assert.Assertions) {
	// File permissions: -rw-r--r--.
	err := ioutil.WriteFile(file.Name(), []byte(content), 0644)
	assert.Nil(err)
}

func TestFillTarantoolCtx(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	var err error
	var ctx = context.Ctx{}

	projectPath, err := ioutil.TempDir("", "project")
	assert.Nil(err)
	defer os.RemoveAll(projectPath)

	// Get Tarantool info from an environment.
	ctxFromEnv := context.Ctx{}
	ctxFromEnv.Project.Path = projectPath
	err = fillTarantoolCtx(&ctxFromEnv)
	assert.Nil(err)
	assert.Equal(ctxFromEnv.Tarantool.IsUserSpecifiedVersion, false)

	// Create tmp tarantool.txt file.
	tarantoolVersionFilePath := filepath.Join(projectPath, "tarantool.txt")
	tarantoolVersionFile, err := os.Create(tarantoolVersionFilePath)
	assert.Nil(err)

	// Tarantool version.
	tarantoolVersionFromFile := "1.2.3-beta1"
	tarantoolVersionSpecified := "3.4.5-rc1"

	// --tarantool-version is specified, tarantool.txt exists.
	ctx = context.Ctx{}
	ctx.Project.Path = projectPath
	ctx.Tarantool.TarantoolVersion = tarantoolVersionSpecified
	writeTarantoolTxtFile(tarantoolVersionFile, fmt.Sprintf("TARANTOOL=%s", tarantoolVersionFromFile),
		assert)

	err = fillTarantoolCtx(&ctx)
	assert.Nil(err)
	assert.Equal(tarantoolVersionSpecified, ctx.Tarantool.TarantoolVersion)
	assert.Equal(ctx.Tarantool.IsUserSpecifiedVersion, true)

	// --tarantool-version isn't specified, tarantool.txt exists.
	ctx = context.Ctx{}
	ctx.Project.Path = projectPath
	err = fillTarantoolCtx(&ctx)
	assert.Nil(err)
	assert.Equal(tarantoolVersionFromFile, ctx.Tarantool.TarantoolVersion)
	assert.Equal(ctx.Tarantool.IsUserSpecifiedVersion, true)

	// Remove tarantool.txt file to get Tarantool info from an environment.
	os.RemoveAll(tarantoolVersionFilePath)
	ctx = context.Ctx{}
	ctx.Project.Path = projectPath

	err = fillTarantoolCtx(&ctx)
	assert.Nil(err)
	assert.Equal(ctxFromEnv.Tarantool.TarantoolVersion, ctx.Tarantool.TarantoolVersion)
	assert.Equal(ctxFromEnv.Tarantool.TarantoolIsEnterprise, ctx.Tarantool.TarantoolIsEnterprise)
	assert.Equal(ctx.Tarantool.IsUserSpecifiedVersion, false)
}
