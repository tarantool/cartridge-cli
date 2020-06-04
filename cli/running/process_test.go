package running

import (
	"testing"

	"github.com/tarantool/cartridge-cli/cli/project"

	"github.com/stretchr/testify/assert"
)

func TestNewInstanceProcess(t *testing.T) {
	assert := assert.New(t)

	ctx := &project.ProjectCtx{}

	ctx.Name = "myapp"
	ctx.Path = "apps/myapp"
	ctx.Entrypoint = "init.lua"
	ctx.ConfPath = "instances.yml"
	ctx.RunDir = "tmp/run"
	ctx.DataDir = "tmp/data"

	process := NewInstanceProcess(ctx, "instance-1")

	assert.Equal("myapp.instance-1", process.ID)
	assert.Equal("apps/myapp/init.lua", process.entrypoint)

	assert.Equal("tmp/data/myapp.instance-1", process.workDir)
	assert.Equal("tmp/run", process.runDir)
	assert.Equal("tmp/run/myapp.instance-1.pid", process.pidFile)

	expEnv := []string{
		"TARANTOOL_APP_NAME=myapp",
		"TARANTOOL_INSTANCE_NAME=instance-1",
		"TARANTOOL_CFG=instances.yml",
		"TARANTOOL_CONSOLE_SOCK=./tmp/run/myapp.instance-1.control",
		"TARANTOOL_PID_FILE=tmp/run/myapp.instance-1.pid",
		"TARANTOOL_WORKDIR=tmp/data/myapp.instance-1",
	}
	assert.ElementsMatch(expEnv, process.env)
}

func TestNewStateboardProcess(t *testing.T) {
	assert := assert.New(t)

	ctx := &project.ProjectCtx{}

	ctx.Name = "myapp"
	ctx.StateboardName = "myapp-stateboard"
	ctx.Path = "apps/myapp"
	ctx.StateboardEntrypoint = "stateboard.init.lua"
	ctx.ConfPath = "instances.yml"
	ctx.RunDir = "tmp/run"
	ctx.DataDir = "tmp/data"

	process := NewStateboardProcess(ctx)

	assert.Equal("myapp-stateboard", process.ID)
	assert.Equal("apps/myapp/stateboard.init.lua", process.entrypoint)

	assert.Equal("tmp/data/myapp-stateboard", process.workDir)
	assert.Equal("tmp/run", process.runDir)
	assert.Equal("tmp/run/myapp-stateboard.pid", process.pidFile)

	expEnv := []string{
		"TARANTOOL_APP_NAME=myapp-stateboard",
		"TARANTOOL_CFG=instances.yml",
		"TARANTOOL_CONSOLE_SOCK=./tmp/run/myapp-stateboard.control",
		"TARANTOOL_PID_FILE=tmp/run/myapp-stateboard.pid",
		"TARANTOOL_WORKDIR=tmp/data/myapp-stateboard",
	}
	assert.ElementsMatch(expEnv, process.env)
}
