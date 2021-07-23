package failover

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func TestBadValidateFailoverSet(t *testing.T) {
	assert := assert.New(t)

	// Specifying invalid mode
	ctx := context.Ctx{}
	ctx.Failover.Mode = "some-invalid-mode"
	_, err := getFailoverOpts(&ctx)
	assert.Equal("Failover mode should be `stateful`, `eventual` or `disabled`", err.Error())

	// Specifying no mode
	ctx = context.Ctx{}
	ctx.Failover.Mode = ""
	ctx.Failover.ParamsJSON = `{"fencing_pause": 4}`
	_, err = getFailoverOpts(&ctx)
	assert.Equal("Failover mode should be `stateful`, `eventual` or `disabled`", err.Error())

	// Eventual mode with with passing state_provider
	ctx = context.Ctx{}
	ctx.Failover.Mode = "eventual"
	ctx.Failover.StateProvider = "stateboard"
	_, err = getFailoverOpts(&ctx)
	assert.Equal("Please, don't specify --state_provider flag when using eventual mode", err.Error())

	// Stateful mode without state_provider
	ctx = context.Ctx{}
	ctx.Failover.Mode = "stateful"
	ctx.Failover.StateProvider = ""
	ctx.Failover.ProviderParamsJSON = `{"uri": "uri", "password": "pass"}`
	_, err = getFailoverOpts(&ctx)
	assert.Equal("Please, specify --state_provider flag when using stateful mode", err.Error())
}
