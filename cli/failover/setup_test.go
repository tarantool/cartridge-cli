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
	/* assert := assert.New(t) */
}

func createYmlFileWithContent(fileName string, content string) error {
	failoverFile, err := os.Create(fileName)
	if err != nil {
		return nil
	}

	_, err = failoverFile.WriteString(content)
	return err
}
