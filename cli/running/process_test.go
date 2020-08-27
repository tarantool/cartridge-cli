package running

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func TestNewInstanceProcess(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	ctx := &context.Ctx{}
	process := &Process{}

	ctx.Project.Name = "myapp"
	ctx.Running.AppDir = "apps/myapp"
	ctx.Running.Entrypoint = "init.lua"
	ctx.Running.ConfPath = "instances.yml"
	ctx.Running.RunDir = "tmp/run"
	ctx.Running.DataDir = "tmp/data"
	ctx.Running.LogDir = "tmp/log"

	process = NewInstanceProcess(ctx, "instance-1")

	assert.Equal("myapp.instance-1", process.ID)
	assert.Equal("apps/myapp/init.lua", process.entrypoint)

	assert.Equal("tmp/data/myapp.instance-1", process.workDir)
	assert.Equal("tmp/run", process.runDir)
	assert.Equal("tmp/run/myapp.instance-1.pid", process.pidFile)
	assert.Equal("tmp/log", process.logDir)
	assert.Equal("tmp/log/myapp.instance-1.log", process.logFile)
	assert.Equal("tmp/run/myapp.instance-1.control", process.consoleSock)

	assert.Equal("tmp/run/myapp.instance-1.notify", process.notifySockPath)

	expEnv := []string{
		"TARANTOOL_APP_NAME=myapp",
		"TARANTOOL_INSTANCE_NAME=instance-1",
		"TARANTOOL_CFG=instances.yml",
		"TARANTOOL_CONSOLE_SOCK=tmp/run/myapp.instance-1.control",
		"TARANTOOL_PID_FILE=tmp/run/myapp.instance-1.pid",
		"TARANTOOL_WORKDIR=tmp/data/myapp.instance-1",
	}
	assert.ElementsMatch(expEnv, process.env)
}

func TestNewStateboardProcess(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	ctx := &context.Ctx{}
	process := &Process{}

	ctx.Project.Name = "myapp"
	ctx.Project.StateboardName = "myapp-stateboard"
	ctx.Running.AppDir = "apps/myapp"
	ctx.Running.StateboardEntrypoint = "stateboard.init.lua"
	ctx.Running.ConfPath = "instances.yml"
	ctx.Running.RunDir = "tmp/run"
	ctx.Running.DataDir = "tmp/data"
	ctx.Running.LogDir = "tmp/log"

	process = NewStateboardProcess(ctx)

	assert.Equal("myapp-stateboard", process.ID)
	assert.Equal("apps/myapp/stateboard.init.lua", process.entrypoint)

	assert.Equal("tmp/data/myapp-stateboard", process.workDir)
	assert.Equal("tmp/run", process.runDir)
	assert.Equal("tmp/run/myapp-stateboard.pid", process.pidFile)
	assert.Equal("tmp/log", process.logDir)
	assert.Equal("tmp/log/myapp-stateboard.log", process.logFile)
	assert.Equal("tmp/run/myapp-stateboard.control", process.consoleSock)

	assert.Equal("tmp/run/myapp-stateboard.notify", process.notifySockPath)

	expEnv := []string{
		"TARANTOOL_APP_NAME=myapp-stateboard",
		"TARANTOOL_CFG=instances.yml",
		"TARANTOOL_CONSOLE_SOCK=tmp/run/myapp-stateboard.control",
		"TARANTOOL_PID_FILE=tmp/run/myapp-stateboard.pid",
		"TARANTOOL_WORKDIR=tmp/data/myapp-stateboard",
	}
	assert.ElementsMatch(expEnv, process.env)
}

func TestPathToEntrypoint(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	ctx := &context.Ctx{}
	process := &Process{}

	// rel path to entrypoint
	ctx.Running.AppDir = "apps/myapp"
	ctx.Running.Entrypoint = "init.lua"
	ctx.Running.StateboardEntrypoint = "stateboard.init.lua"

	process = NewInstanceProcess(ctx, "instance-1")
	assert.Equal("apps/myapp/init.lua", process.entrypoint)

	process = NewStateboardProcess(ctx)
	assert.Equal("apps/myapp/stateboard.init.lua", process.entrypoint)

	// abs path to entrypoint
	ctx.Running.AppDir = "apps/myapp"
	ctx.Running.Entrypoint = "/abs/path/to/init.lua"
	ctx.Running.StateboardEntrypoint = "/abs/path/to/stateboard.init.lua"

	process = NewInstanceProcess(ctx, "instance-1")
	assert.Equal("/abs/path/to/init.lua", process.entrypoint)

	process = NewStateboardProcess(ctx)
	assert.Equal("/abs/path/to/stateboard.init.lua", process.entrypoint)
}
