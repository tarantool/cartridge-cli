package running

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func TestFillCtx(t *testing.T) {
	assert := assert.New(t)

	var ctx *context.Ctx

	const APPNAME = "myapp"

	nonExistentConf := "non-existent-path"

	var err error
	var args []string

	// local, no args, app name is specified
	ctx = &context.Ctx{}
	ctx.Project.Name = APPNAME
	ctx.Running.Global = false
	ctx.Running.ConfPath = nonExistentConf
	args = []string{}

	err = FillCtx(ctx, args)
	assert.Nil(err)
	assert.Equal(fmt.Sprintf("%s-stateboard", APPNAME), ctx.Project.StateboardName)
	assert.Nil(ctx.Running.Instances)

	// local, instances are specified, app name is specified
	ctx = &context.Ctx{}
	ctx.Project.Name = APPNAME
	ctx.Running.Global = false
	ctx.Running.ConfPath = nonExistentConf
	args = []string{"router", "s1-master", "s1-replica"}

	err = FillCtx(ctx, args)
	assert.Nil(err)
	assert.Equal(fmt.Sprintf("%s-stateboard", APPNAME), ctx.Project.StateboardName)
	assert.Equal(args, ctx.Running.Instances)

	// global, no args, app name isn't specified
	ctx = &context.Ctx{}
	ctx.Running.Global = true
	ctx.Running.ConfPath = nonExistentConf
	args = []string{}

	err = FillCtx(ctx, args)
	assert.True(strings.Contains(err.Error(), "APP_NAME or --name should be specified"))

	// global, no args, app name is specified by --name
	ctx = &context.Ctx{}
	ctx.Project.Name = APPNAME
	ctx.Running.Global = true
	ctx.Running.ConfPath = nonExistentConf
	args = []string{}

	err = FillCtx(ctx, args)
	assert.Nil(err)
	assert.Equal(fmt.Sprintf("%s-stateboard", APPNAME), ctx.Project.StateboardName)
	assert.Nil(ctx.Running.Instances)

	// global, instances are specified, app name is specified by --name
	ctx = &context.Ctx{}
	ctx.Project.Name = APPNAME
	ctx.Running.Global = true
	ctx.Running.ConfPath = nonExistentConf
	args = []string{"router", "s1-master", "s1-replica"}

	err = FillCtx(ctx, args)
	assert.Nil(err)
	assert.Equal(fmt.Sprintf("%s-stateboard", APPNAME), ctx.Project.StateboardName)
	assert.Equal(args, ctx.Running.Instances)

	// global, instances and app name are specified in args
	ctx = &context.Ctx{}
	ctx.Running.Global = true
	ctx.Running.ConfPath = nonExistentConf
	args = []string{APPNAME, "router", "s1-master", "s1-replica"}

	err = FillCtx(ctx, args)
	assert.Nil(err)
	assert.Equal(APPNAME, ctx.Project.Name)
	assert.Equal(fmt.Sprintf("%s-stateboard", APPNAME), ctx.Project.StateboardName)
	assert.Equal(args[1:], ctx.Running.Instances)

	// global, only app name is specified in args
	ctx = &context.Ctx{}
	ctx.Running.Global = true
	ctx.Running.ConfPath = nonExistentConf
	args = []string{APPNAME}

	err = FillCtx(ctx, args)
	assert.Nil(err)
	assert.Equal(APPNAME, ctx.Project.Name)
	assert.Equal(fmt.Sprintf("%s-stateboard", APPNAME), ctx.Project.StateboardName)
	assert.Nil(ctx.Running.Instances)

	// local, stateboard only is set
	ctx = &context.Ctx{}
	ctx.Project.Name = APPNAME
	ctx.Running.Global = false
	ctx.Running.ConfPath = nonExistentConf
	ctx.Running.StateboardOnly = true
	args = []string{"router", "s1-master", "s1-replica"}

	err = FillCtx(ctx, args)
	assert.Nil(err)
	assert.Equal(APPNAME, ctx.Project.Name)
	assert.True(ctx.Running.WithStateboard)
	assert.Equal(fmt.Sprintf("%s-stateboard", APPNAME), ctx.Project.StateboardName)
	assert.Nil(ctx.Running.Instances)

	// global, stateboard only is set, app name is specified via --name
	ctx = &context.Ctx{}
	ctx.Project.Name = APPNAME
	ctx.Running.Global = true
	ctx.Running.ConfPath = nonExistentConf
	ctx.Running.StateboardOnly = true
	args = []string{}

	err = FillCtx(ctx, args)
	assert.Nil(err)
	assert.Equal(APPNAME, ctx.Project.Name)
	assert.True(ctx.Running.WithStateboard)
	assert.Equal(fmt.Sprintf("%s-stateboard", APPNAME), ctx.Project.StateboardName)
	assert.Nil(ctx.Running.Instances)

	// global, stateboard only is set, app name is specified as an arg
	ctx = &context.Ctx{}
	ctx.Running.Global = true
	ctx.Running.ConfPath = nonExistentConf
	ctx.Running.StateboardOnly = true
	args = []string{APPNAME}

	err = FillCtx(ctx, args)
	assert.Nil(err)
	assert.Equal(APPNAME, ctx.Project.Name)
	assert.True(ctx.Running.WithStateboard)
	assert.Equal(fmt.Sprintf("%s-stateboard", APPNAME), ctx.Project.StateboardName)
	assert.Nil(ctx.Running.Instances)

	// global, stateboard only is set, app name and instances are specified in args
	ctx = &context.Ctx{}
	ctx.Running.Global = true
	ctx.Running.ConfPath = nonExistentConf
	ctx.Running.StateboardOnly = true
	args = []string{APPNAME, "router", "s1-master", "s1-replica"}

	err = FillCtx(ctx, args)
	assert.Nil(err)
	assert.Equal(APPNAME, ctx.Project.Name)
	assert.True(ctx.Running.WithStateboard)
	assert.Equal(fmt.Sprintf("%s-stateboard", APPNAME), ctx.Project.StateboardName)
	assert.Nil(ctx.Running.Instances)
}
