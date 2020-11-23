package project

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/stretchr/testify/assert"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func TestGetPath(t *testing.T) {
	assert := assert.New(t)

	var err error
	var path string
	var conf map[string]interface{}

	const specifiedPath = "specifiedPath"
	const defaultPath = "defaultPath"
	const sectionName = "sectionName"
	const otherSectionName = "otherSectionName"
	const sectionValue = "sectionValue"
	const basePath = "basePath"

	// path is specified
	path, err = getPath(nil, PathOpts{
		SpecifiedPath: specifiedPath,
		DefaultPath:   defaultPath,
	})
	assert.Nil(err)
	assert.Equal(specifiedPath, path)

	// path is specified, base path is specified
	path, err = getPath(nil, PathOpts{
		SpecifiedPath: specifiedPath,
		DefaultPath:   defaultPath,
		BasePath:      basePath,
	})
	assert.Nil(err)
	assert.Equal(filepath.Join(basePath, specifiedPath), path)

	// path isn't specified
	path, err = getPath(nil, PathOpts{
		DefaultPath: defaultPath,
	})
	assert.Nil(err)
	assert.Equal(defaultPath, path)

	// path isn't specified, base path is specified
	path, err = getPath(nil, PathOpts{
		DefaultPath: defaultPath,
		BasePath:    basePath,
	})
	assert.Nil(err)
	assert.Equal(filepath.Join(basePath, defaultPath), path)

	// path isn't specified, defaultPath is empty, base path is specified
	path, err = getPath(nil, PathOpts{
		DefaultPath: "",
		BasePath:    basePath,
	})
	assert.Nil(err)
	assert.Equal("", path)

	// specified conf, but no section
	conf = map[string]interface{}{
		sectionName: sectionValue,
	}
	path, err = getPath(conf, PathOpts{
		DefaultPath: defaultPath,
	})
	assert.Nil(err)
	assert.Equal(defaultPath, path)

	// specified section, but no conf
	path, err = getPath(nil, PathOpts{
		DefaultPath:     defaultPath,
		ConfSectionName: sectionName,
	})
	assert.Nil(err)
	assert.Equal(defaultPath, path)

	// specified section not present in conf
	conf = map[string]interface{}{
		otherSectionName: sectionValue,
	}
	path, err = getPath(conf, PathOpts{
		DefaultPath:     defaultPath,
		ConfSectionName: sectionName,
	})
	assert.Nil(err)
	assert.Equal(defaultPath, path)

	// specified section present in conf
	conf = map[string]interface{}{
		sectionName: sectionValue,
	}
	path, err = getPath(conf, PathOpts{
		DefaultPath:     defaultPath,
		ConfSectionName: sectionName,
	})
	assert.Nil(err)
	assert.Equal(sectionValue, path)

	// specified section present in conf, base path is specified
	conf = map[string]interface{}{
		sectionName: sectionValue,
	}
	path, err = getPath(conf, PathOpts{
		DefaultPath:     defaultPath,
		ConfSectionName: sectionName,
		BasePath:        basePath,
	})
	assert.Nil(err)
	assert.Equal(filepath.Join(basePath, sectionValue), path)

	// specified section present in conf with no string value
	conf = map[string]interface{}{
		sectionName: true,
	}
	path, err = getPath(conf, PathOpts{
		DefaultPath:     defaultPath,
		ConfSectionName: sectionName,
	})
	assert.True(strings.Contains(err.Error(), "config value should be string"))
}

func TestSetLocalRunningPathsDefault(t *testing.T) {
	assert := assert.New(t)

	var err error
	var ctx *context.Ctx

	homeDir, err := common.GetHomeDir()
	if err != nil {
		homeDir = defaultHomeDir
	}

	defaultCartridgeTmpDir = filepath.Join(homeDir, defaultCartridgeTmpDirName)

	// create tmp app directory
	dir, err := ioutil.TempDir("", "myapp")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// CartridgeTmpDir isn't specified

	ctx = &context.Ctx{}
	ctx.Running.AppDir = dir
	var runDir string

	// no flags or conf sections specified

	err = SetLocalRunningPaths(ctx)
	assert.Nil(err)

	assert.True(ctx.Project.ID != "")

	assert.Equal(filepath.Join(ctx.Running.AppDir, defaultLocalConfPath), ctx.Running.ConfPath)
	assert.Equal(filepath.Join(ctx.Running.AppDir, defaultLocalDataDir), ctx.Running.DataDir)
	assert.Equal(filepath.Join(ctx.Running.AppDir, defaultLocalLogDir), ctx.Running.LogDir)
	assert.Equal(filepath.Join(ctx.Running.AppDir, defaultEntrypoint), ctx.Running.Entrypoint)
	assert.Equal(filepath.Join(ctx.Running.AppDir, defaultStateboardEntrypoint), ctx.Running.StateboardEntrypoint)

	runDir = filepath.Join(homeDir, defaultCartridgeTmpDirName, fmt.Sprintf("run-%s", ctx.Project.ID))
	assert.Equal(runDir, ctx.Running.RunDir)

	// CartridgeTmpDir is specified

	ctx = &context.Ctx{}
	ctx.Running.AppDir = dir
	ctx.Cli.CartridgeTmpDir = "cartridge-tempdir-from-env"

	// no flags or conf sections specified

	err = SetLocalRunningPaths(ctx)
	assert.Nil(err)

	assert.True(ctx.Project.ID != "")

	assert.Equal(filepath.Join(ctx.Running.AppDir, defaultLocalConfPath), ctx.Running.ConfPath)
	assert.Equal(filepath.Join(ctx.Running.AppDir, defaultLocalDataDir), ctx.Running.DataDir)
	assert.Equal(filepath.Join(ctx.Running.AppDir, defaultLocalLogDir), ctx.Running.LogDir)
	assert.Equal(filepath.Join(ctx.Running.AppDir, defaultEntrypoint), ctx.Running.Entrypoint)
	assert.Equal(filepath.Join(ctx.Running.AppDir, defaultStateboardEntrypoint), ctx.Running.StateboardEntrypoint)

	runDir = filepath.Join(ctx.Cli.CartridgeTmpDir, fmt.Sprintf("run-%s", ctx.Project.ID))
	assert.Equal(runDir, ctx.Running.RunDir)
}

func TestSetLocalRunningPathsSpecified(t *testing.T) {
	assert := assert.New(t)

	var err error

	homeDir, err := common.GetHomeDir()
	if err != nil {
		homeDir = defaultHomeDir
	}

	defaultCartridgeTmpDir = filepath.Join(homeDir, defaultCartridgeTmpDirName)

	// create tmp app directory
	dir, err := ioutil.TempDir("", "myapp")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	ctx := &context.Ctx{}
	ctx.Running.AppDir = dir

	ctx.Running.ConfPath = "conf"
	ctx.Running.DataDir = "data-dir"
	ctx.Running.LogDir = "log-dir"
	ctx.Running.Entrypoint = "entrypoint"
	ctx.Running.StateboardEntrypoint = "stateboard-entrypoint"

	// no flags or conf sections specified

	err = SetLocalRunningPaths(ctx)
	assert.Nil(err)

	assert.True(ctx.Project.ID != "")

	assert.Equal(ctx.Running.ConfPath, ctx.Running.ConfPath)
	assert.Equal(ctx.Running.DataDir, ctx.Running.DataDir)
	assert.Equal(ctx.Running.LogDir, ctx.Running.LogDir)
	assert.Equal(ctx.Running.RunDir, ctx.Running.RunDir)
	assert.Equal(ctx.Running.Entrypoint, ctx.Running.Entrypoint)
	assert.Equal(ctx.Running.StateboardEntrypoint, ctx.Running.StateboardEntrypoint)
}

func TestSetLocalRunningPathsFromConf(t *testing.T) {
	assert := assert.New(t)

	var err error

	// create tmp app directory
	dir, err := ioutil.TempDir("", "myapp")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	ctx := &context.Ctx{}
	ctx.Running.AppDir = dir

	// all paths are specified in conf
	conf := map[string]string{
		"cfg":      "cfg-from-conf",
		"data-dir": "data-dir-from-conf",
		"run-dir":  "run-dir-from-conf",
		"log-dir":  "log-dir-from-conf",
		"script":   "script-from-conf",
	}
	cartridgeConfPath := filepath.Join(ctx.Running.AppDir, cartridgeLocalConf)
	writeCliConf(cartridgeConfPath, conf)

	err = SetLocalRunningPaths(ctx)
	assert.Nil(err)

	assert.Equal(filepath.Join(ctx.Running.AppDir, conf["cfg"]), ctx.Running.ConfPath)
	assert.Equal(filepath.Join(ctx.Running.AppDir, conf["data-dir"]), ctx.Running.DataDir)
	assert.Equal(filepath.Join(ctx.Running.AppDir, conf["run-dir"]), ctx.Running.RunDir)
	assert.Equal(filepath.Join(ctx.Running.AppDir, conf["log-dir"]), ctx.Running.LogDir)
	assert.Equal(filepath.Join(ctx.Running.AppDir, conf["script"]), ctx.Running.Entrypoint)
	assert.Equal(filepath.Join(ctx.Running.AppDir, defaultStateboardEntrypoint), ctx.Running.StateboardEntrypoint)
}

func writeCliConf(confPath string, conf map[string]string) {
	file, err := os.Create(confPath)
	if err != nil {
		log.Fatal(err)
	}

	contentBytes, err := yaml.Marshal(conf)
	if err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile(file.Name(), contentBytes, 0644); err != nil {
		log.Fatalf("Failed to write config: %s", err)
	}
}
