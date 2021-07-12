package failover

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func TestParseFailoverYMLFIlePositive(t *testing.T) {
	assert := assert.New(t)

	// Eventual failover test parsing
	ctx := context.Ctx{}
	ctx.Failover.File = "failover_test_eventual"
	err := createYmlFileWithContent(ctx.Failover.File, `
mode: eventual
failover_timeout: 1
fencing_enabled: true
fencing_timeout: 88
fencing_pause: 4`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	err = parseFailoverParams(&ctx)
	assert.Equal(nil, err)
	assert.Equal(context.FailoverCtx{
		File:            "failover_test_eventual",
		Mode:            "eventual",
		FailoverTimeout: 1,
		FencingEnabled:  true,
		FencingTimeout:  88,
		FencingPause:    4,
	}, ctx.Failover)

	// Stateful stateboard failover test parsing
	ctx = context.Ctx{}
	ctx.Failover.File = "failover_test_stateboard"
	err = createYmlFileWithContent(ctx.Failover.File, `
mode: stateful
state_provider: stateboard
stateboard_params:
  uri: yuriy
  password: stroganov-bmstu
fencing_enabled: false
fencing_timeout: 380`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	err = parseFailoverParams(&ctx)
	assert.Equal(nil, err)
	assert.Equal(context.FailoverCtx{
		File:           "failover_test_stateboard",
		Mode:           "stateful",
		StateProvider:  "stateboard",
		FencingEnabled: false,
		FencingTimeout: 380,
		StateboardParams: context.StateboardParamsCtx{
			URI:      "yuriy",
			Password: "stroganov-bmstu",
		},
	}, ctx.Failover)

	// Stateful etcd2 failover test parsing
	ctx = context.Ctx{}
	ctx.Failover.File = "failover_test_etcd2"
	err = createYmlFileWithContent(ctx.Failover.File, `
mode: stateful
state_provider: etcd2
etcd2_params:
  prefix: xiferp
  lock_delay: 120
  endpoints: [http://localhost:2379, http://localhost:4001]
  password: superpass
  username: superuser`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	err = parseFailoverParams(&ctx)
	assert.Equal(nil, err)
	assert.Equal(context.FailoverCtx{
		File:          "failover_test_etcd2",
		Mode:          "stateful",
		StateProvider: "etcd2",
		Etcd2Params: context.Etcd2ParamsCtx{
			Prefix:    "xiferp",
			Password:  "superpass",
			Username:  "superuser",
			LockDelay: 120,
			Endpoints: []string{"http://localhost:2379", "http://localhost:4001"},
		},
	}, ctx.Failover)
}

func TestParseFailoverYMLFIleNegative(t *testing.T) {
	assert := assert.New(t)

	// Specifying invalid mode
	ctx := context.Ctx{}
	ctx.Failover.File = "failover_test_invalid_mode_1"
	err := createYmlFileWithContent(ctx.Failover.File, `mode: some-invalid-mode`)
	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	// Specifying no mode
	ctx = context.Ctx{}
	ctx.Failover.File = "failover_test_invalid_mode_2"
	err = createYmlFileWithContent(ctx.Failover.File, `fencing_pause: 4`)
	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	err = parseFailoverParams(&ctx)
	assert.Equal("Failover mode should be `stateful` or `eventual`", err.Error())

	// Eventual mode with incorrect params
	// Passing state_provider
	ctx = context.Ctx{}
	ctx.Failover.File = "failover_test_eventual_1"
	err = createYmlFileWithContent(ctx.Failover.File, `
mode: eventual
state_provider: stateboard`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	err = parseFailoverParams(&ctx)
	assert.Equal("You don't have to specify `state_provider` when using eventual mode", err.Error())

	// Passing stateboard_params
	ctx = context.Ctx{}
	ctx.Failover.File = "failover_test_eventual_2"
	err = createYmlFileWithContent(ctx.Failover.File, `
mode: eventual
stateboard_params:
  uri: uri
  password: pass`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	err = parseFailoverParams(&ctx)
	assert.Equal("You don't have to specify `stateboard_params` when using eventual mode", err.Error())

	// Passing etcd2_params
	ctx = context.Ctx{}
	ctx.Failover.File = "failover_test_eventual_3"
	err = createYmlFileWithContent(ctx.Failover.File, `
mode: eventual
etcd2_params:
  uri: uri
  password: pass`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	err = parseFailoverParams(&ctx)
	assert.Equal("You don't have to specify `etcd2_params` when using eventual mode", err.Error())

	// Stateful with incorrect params
	// No state_provider
	ctx = context.Ctx{}
	ctx.Failover.File = "failover_test_stateful_stateboard_1"
	err = createYmlFileWithContent(ctx.Failover.File, `
mode: stateful
stateboard_params:
  uri: uri
  password: pass`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	err = parseFailoverParams(&ctx)
	assert.Equal("Failover `state_provider` should be `stateboard` or `etcd2`", err.Error())

	// No stateboard_params
	ctx = context.Ctx{}
	ctx.Failover.File = "failover_test_stateful_stateboard_2"
	err = createYmlFileWithContent(ctx.Failover.File, `
mode: stateful
state_provider: stateboard
etcd2_params:
  uri: uri
  password: pass`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	err = parseFailoverParams(&ctx)
	assert.Equal("You should specify `stateboard_params` when using stateboard provider", err.Error())

	// Stateful stateboard and etcd2_params
	ctx = context.Ctx{}
	ctx.Failover.File = "failover_test_stateful_stateboard_3"
	err = createYmlFileWithContent(ctx.Failover.File, `
mode: stateful
state_provider: stateboard
stateboard_params:
  uri: uri
  password: pass
etcd2_params:
  uri: uri
  password: pass`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	err = parseFailoverParams(&ctx)
	assert.Equal("You shouldn't specify `etcd2_params` when using stateboard provider", err.Error())

	// No password in stateboard params
	ctx = context.Ctx{}
	ctx.Failover.File = "failover_test_stateful_stateboard_4"
	err = createYmlFileWithContent(ctx.Failover.File, `
mode: stateful
state_provider: stateboard
stateboard_params:
  uri: uri`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	err = parseFailoverParams(&ctx)
	assert.Equal("You should specify `uri` and `password` params when using stateboard provider", err.Error())

	// Etcd2 provider with stateboard_params
	ctx = context.Ctx{}
	ctx.Failover.File = "failover_test_stateful_etcd2_1"
	err = createYmlFileWithContent(ctx.Failover.File, `
mode: stateful
state_provider: etcd2
stateboard_params:
  uri: uri`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	err = parseFailoverParams(&ctx)
	assert.Equal("You should specify `etcd2_params` when using stateboard provider", err.Error())

	// Etcd2 provider with etcd2_params
	ctx = context.Ctx{}
	ctx.Failover.File = "failover_test_stateful_etcd2_2"
	err = createYmlFileWithContent(ctx.Failover.File, `
mode: stateful
state_provider: etcd2
etcd2_params:
  lock_delay: 123
stateboard_params:
  uri: uri
  password: pass`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	err = parseFailoverParams(&ctx)
	assert.Equal("You shouldn't specify `stateboard_params` when using etcd2 provider", err.Error())

	// Negative lock_delay
	ctx = context.Ctx{}
	ctx.Failover.File = "failover_test_stateful_etcd2_3"
	err = createYmlFileWithContent(ctx.Failover.File, `
mode: stateful
state_provider: etcd2
etcd2_params:
  lock_delay: -100500`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	err = parseFailoverParams(&ctx)
	assert.Equal("Parameter lock_delay must be greater than or equal to 0", err.Error())

	// Negative failover_timeout
	ctx = context.Ctx{}
	ctx.Failover.File = "failover_test_negative_1"
	err = createYmlFileWithContent(ctx.Failover.File, `
mode: eventual
failover_timeout: -10`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	err = parseFailoverParams(&ctx)
	assert.Equal("Parameter failover_timeout must be greater than or equal to 0", err.Error())

	// Negative fencing timeout
	ctx = context.Ctx{}
	ctx.Failover.File = "failover_test_negative_2"
	err = createYmlFileWithContent(ctx.Failover.File, `
mode: eventual
failover_timeout: 10
fencing_timeout: -200`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	err = parseFailoverParams(&ctx)
	assert.Equal("Parameter fencing_timeout must be greater than or equal to 0", err.Error())

	// Negative fencing pause
	ctx = context.Ctx{}
	ctx.Failover.File = "failover_test_negative_3"
	err = createYmlFileWithContent(ctx.Failover.File, `
mode: eventual
fencing_pause: -500`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	err = parseFailoverParams(&ctx)
	assert.Equal("Parameter fencing_pause must be greater than or equal to 0", err.Error())
}

func createYmlFileWithContent(fileName string, content string) error {
	failoverFile, err := os.Create(fileName)
	if err != nil {
		return nil
	}

	_, err = failoverFile.WriteString(content)
	return err
}
