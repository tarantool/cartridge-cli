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

	curDir, err := os.Getwd()
	assert.Nil(err)

	const specifiedPath = "specifiedPath"
	const defaultPath = "defaultPath"
	const sectionName = "sectionName"
	const otherSectionName = "otherSectionName"
	const sectionValue = "sectionValue"

	// path is specified
	path, err = getPath(nil, PathOpts{
		SpecifiedPath: specifiedPath,
		DefaultPath:   defaultPath,
	})
	assert.Nil(err)
	assert.Equal(specifiedPath, path)

	// path is specified, GetAbs
	path, err = getPath(nil, PathOpts{
		SpecifiedPath: specifiedPath,
		DefaultPath:   defaultPath,
		GetAbs:        true,
	})
	assert.Nil(err)
	assert.Equal(filepath.Join(curDir, specifiedPath), path)

	// path isn't specified
	path, err = getPath(nil, PathOpts{
		DefaultPath: defaultPath,
	})
	assert.Nil(err)
	assert.Equal(defaultPath, path)

	// path isn't specified, GetAbs
	path, err = getPath(nil, PathOpts{
		DefaultPath: defaultPath,
		GetAbs:      true,
	})
	assert.Nil(err)
	assert.Equal(filepath.Join(curDir, defaultPath), path)

	// path isn't specified, defaultPath is empty, GetAbs
	path, err = getPath(nil, PathOpts{
		DefaultPath: "",
		GetAbs:      true,
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

	// specified section present in conf, GetAbs
	conf = map[string]interface{}{
		sectionName: sectionValue,
	}
	path, err = getPath(conf, PathOpts{
		DefaultPath:     defaultPath,
		ConfSectionName: sectionName,
		GetAbs:          true,
	})
	assert.Nil(err)
	assert.Equal(filepath.Join(curDir, sectionValue), path)

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

func TestSetRunningConfPath(t *testing.T) {
	assert := assert.New(t)

	var err error

	curDir, err := os.Getwd()
	assert.Nil(err)

	homeDir, err := common.GetHomeDir()
	assert.Nil(err)

	var ctx = &context.Ctx{}

	// local
	ctx.Running.Global = false
	ctx.Running.ConfPath = ""
	err = setRunningConfPath(ctx)
	assert.Nil(err)
	assert.Equal(filepath.Join(curDir, runningConfFilename), ctx.Running.ConfPath)

	// global
	ctx.Running.Global = true
	ctx.Running.ConfPath = ""
	err = setRunningConfPath(ctx)
	assert.Nil(err)
	assert.Equal(filepath.Join(homeDir, runningConfFilename), ctx.Running.ConfPath)

	// already set
	ctx.Running.Global = false
	ctx.Running.ConfPath = "/already/set"
	err = setRunningConfPath(ctx)
	assert.Nil(err)
	assert.Equal("/already/set", ctx.Running.ConfPath)
}

func writeConf(file *os.File, conf interface{}) {
	content, err := yaml.Marshal(conf)
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(file.Name(), content, 0644); err != nil {
		panic(fmt.Errorf("Failed to write conf: %s", err))
	}
}

func abs(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return absPath
}

func TestSetRunningPaths(t *testing.T) {
	assert := assert.New(t)

	var err error

	// create tmp config file
	f, err := ioutil.TempFile("", ".cartridge.yml")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	curDir, err := os.Getwd()
	assert.Nil(err)

	var ctx *context.Ctx
	var conf RunningConf

	conf = RunningConf{
		"cfg":      "cfg-from-conf",
		"run-dir":  "run-dir-from-conf",
		"data-dir": "data-dir-from-conf",
		"log-dir":  "log-dir-from-conf",
		"apps-dir": "apps-dir-from-conf",
		"script":   "script-from-conf",
	}
	writeConf(f, conf)

	const APPNAME = "myapp"

	// local, useConf is false
	// nothing is specified by user
	// expected: default values
	ctx = &context.Ctx{}
	ctx.Project.Name = APPNAME
	ctx.Running.Global = false
	ctx.Running.ConfPath = f.Name()

	err = SetRunningPaths(ctx, false)
	assert.Nil(err)

	assert.Equal(abs("instances.yml"), ctx.Running.AppConfPath)
	assert.Equal(abs("tmp/run"), ctx.Running.RunDir)
	assert.Equal(abs("tmp/data"), ctx.Running.DataDir)
	assert.Equal(abs("tmp/log"), ctx.Running.LogDir)
	assert.Equal("", ctx.Running.AppsDir)
	assert.Equal(curDir, ctx.Running.AppDir)
	assert.Equal("init.lua", ctx.Running.Entrypoint)
	assert.Equal("stateboard.init.lua", ctx.Running.StateboardEntrypoint)

	// global, useConf is false
	// nothing is specified by user
	// expected: default values
	ctx = &context.Ctx{}
	ctx.Project.Name = APPNAME
	ctx.Running.Global = true
	ctx.Running.ConfPath = f.Name()

	err = SetRunningPaths(ctx, false)
	assert.Nil(err)

	assert.Equal("/etc/tarantool/conf.d", ctx.Running.AppConfPath)
	assert.Equal("/var/run/tarantool", ctx.Running.RunDir)
	assert.Equal("/var/lib/tarantool", ctx.Running.DataDir)
	assert.Equal("/var/log/tarantool", ctx.Running.LogDir)
	assert.Equal("/usr/share/tarantool", ctx.Running.AppsDir)
	assert.Equal(filepath.Join("/usr/share/tarantool", APPNAME), ctx.Running.AppDir)
	assert.Equal("init.lua", ctx.Running.Entrypoint)
	assert.Equal("stateboard.init.lua", ctx.Running.StateboardEntrypoint)

	// local, useConf is true
	// nothing is specified by user
	// expected: values from conf
	ctx = &context.Ctx{}
	ctx.Project.Name = APPNAME
	ctx.Running.Global = false
	ctx.Running.ConfPath = f.Name()

	err = SetRunningPaths(ctx, true)
	assert.Nil(err)

	assert.Equal(abs("cfg-from-conf"), ctx.Running.AppConfPath)
	assert.Equal(abs("run-dir-from-conf"), ctx.Running.RunDir)
	assert.Equal(abs("data-dir-from-conf"), ctx.Running.DataDir)
	assert.Equal(abs("log-dir-from-conf"), ctx.Running.LogDir)
	assert.Equal("", ctx.Running.AppsDir)
	assert.Equal(curDir, ctx.Running.AppDir)
	assert.Equal("script-from-conf", ctx.Running.Entrypoint)
	assert.Equal("stateboard.init.lua", ctx.Running.StateboardEntrypoint)

	// global, useConf is true
	// nothing is specified by user
	// expected: values from conf
	ctx = &context.Ctx{}
	ctx.Project.Name = APPNAME
	ctx.Running.Global = true
	ctx.Running.ConfPath = f.Name()

	err = SetRunningPaths(ctx, true)
	assert.Nil(err)

	assert.Equal(abs("cfg-from-conf"), ctx.Running.AppConfPath)
	assert.Equal(abs("run-dir-from-conf"), ctx.Running.RunDir)
	assert.Equal(abs("data-dir-from-conf"), ctx.Running.DataDir)
	assert.Equal(abs("log-dir-from-conf"), ctx.Running.LogDir)
	assert.Equal(abs("apps-dir-from-conf"), ctx.Running.AppsDir)
	assert.Equal(filepath.Join(abs("apps-dir-from-conf"), APPNAME), ctx.Running.AppDir)
	assert.Equal("script-from-conf", ctx.Running.Entrypoint)
	assert.Equal("stateboard.init.lua", ctx.Running.StateboardEntrypoint)

	// local, useConf is true
	// values are specified by user
	// expected: values by user (some of them become abs)
	ctx = &context.Ctx{}
	ctx.Project.Name = APPNAME
	ctx.Running.Global = false
	ctx.Running.ConfPath = f.Name()

	ctx.Running.AppConfPath = "cfg-by-user"
	ctx.Running.RunDir = "run-dir-by-user"
	ctx.Running.DataDir = "data-dir-by-user"
	ctx.Running.LogDir = "log-dir-by-user"
	ctx.Running.Entrypoint = "script-by-user"

	err = SetRunningPaths(ctx, true)
	assert.Nil(err)

	assert.Equal(abs("cfg-by-user"), ctx.Running.AppConfPath)
	assert.Equal(abs("run-dir-by-user"), ctx.Running.RunDir)
	assert.Equal(abs("data-dir-by-user"), ctx.Running.DataDir)
	assert.Equal(abs("log-dir-by-user"), ctx.Running.LogDir)
	assert.Equal("", ctx.Running.AppsDir)
	assert.Equal(curDir, ctx.Running.AppDir)
	assert.Equal("script-by-user", ctx.Running.Entrypoint)
	assert.Equal("stateboard.init.lua", ctx.Running.StateboardEntrypoint)

	// global, useConf is true
	// values are specified by user
	// expected: values by user (some of them become abs)
	ctx = &context.Ctx{}
	ctx.Project.Name = APPNAME
	ctx.Running.Global = true
	ctx.Running.ConfPath = f.Name()

	ctx.Running.AppConfPath = "cfg-by-user"
	ctx.Running.RunDir = "run-dir-by-user"
	ctx.Running.DataDir = "data-dir-by-user"
	ctx.Running.LogDir = "log-dir-by-user"
	ctx.Running.AppsDir = "apps-dir-by-user"
	ctx.Running.Entrypoint = "script-by-user"

	err = SetRunningPaths(ctx, true)
	assert.Nil(err)

	assert.Equal(abs("cfg-by-user"), ctx.Running.AppConfPath)
	assert.Equal(abs("run-dir-by-user"), ctx.Running.RunDir)
	assert.Equal(abs("data-dir-by-user"), ctx.Running.DataDir)
	assert.Equal(abs("log-dir-by-user"), ctx.Running.LogDir)
	assert.Equal(abs("apps-dir-by-user"), ctx.Running.AppsDir)
	assert.Equal(filepath.Join(abs("apps-dir-by-user"), APPNAME), ctx.Running.AppDir)
	assert.Equal("script-by-user", ctx.Running.Entrypoint)
	assert.Equal("stateboard.init.lua", ctx.Running.StateboardEntrypoint)
}
