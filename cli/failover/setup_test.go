package failover

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func TestParseFailoverYMLFile(t *testing.T) {
	assert := assert.New(t)

	// Disabled mode
	ctx := context.Ctx{}
	ctx.Failover.File = "failover_test_disabled"
	err := createYmlFileWithContent(ctx.Failover.File, `mode: disabled`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	opts, err := getFailoverOptsFromFile(&ctx)
	assert.Equal(nil, err)
	assert.Equal(&FailoverOpts{"mode": "disabled"}, opts)

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

	opts, err = getFailoverOptsFromFile(&ctx)
	assert.Equal(nil, err)

	failoverTimeout, fencingTimeout, fencingPause := 1, 88, 4
	fencingFlag := true

	assert.Equal(&FailoverOpts{
		"mode":             "eventual",
		"failover_timeout": 1,
		"fencing_enabled":  true,
		"fencing_timeout":  88,
		"fencing_pause":    4,
	}, opts)

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

	opts, err = getFailoverOptsFromFile(&ctx)
	assert.Equal(nil, err)

	provider := "stateboard"
	fencingFlag = false
	fencingTimeout = 380

	assert.Equal(&FailoverOpts{
		"mode":            "stateful",
		"state_provider":  "stateboard",
		"fencing_enabled": false,
		"fencing_timeout": 380,
		"stateboard_params": map[interface{}]interface{}{
			"uri":      "yuriy",
			"password": "stroganov-bmstu",
		},
	}, opts)

	// Stateful etcd2 failover test parsing
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

	opts, err = getFailoverOptsFromFile(&ctx)
	assert.Equal(nil, err)

	provider = "etcd2"

	assert.Equal(&FailoverOpts{
		"mode":           "stateful",
		"state_provider": "etcd2",
		"etcd2_params": map[interface{}]interface{}{
			"prefix":     "xiferp",
			"password":   "superpass",
			"username":   "superuser",
			"lock_delay": 120,
			"endpoints":  []interface{}{"http://localhost:2379", "http://localhost:4001"},
		},
	}, opts)
}

func TestGoodValidateFailoverYMLFile(t *testing.T) {
	assert := assert.New(t)

	ctx := context.Ctx{}
	// Stateful stateboard and etcd2_params
	ctx.Failover.File = "failover_validate_1"
	err := createYmlFileWithContent(ctx.Failover.File, `
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

	_, err = getFailoverOptsFromFile(&ctx)
	assert.Equal(nil, err)

	// Stateful etcd2 no etcd2_params
	ctx.Failover.File = "failover_validate_2"
	err = createYmlFileWithContent(ctx.Failover.File, `
mode: stateful
state_provider: etcd2
stateboard_params:
  uri: uri`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	_, err = getFailoverOptsFromFile(&ctx)
	assert.Equal(nil, err)

	// Stateful etcd2 failover with stateboard_params
	ctx.Failover.File = "failover_validate_3"
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

	_, err = getFailoverOptsFromFile(&ctx)
	assert.Equal(nil, err)

	ctx = context.Ctx{}
	ctx.Failover.File = "failover_validate_4"
	err = createYmlFileWithContent(ctx.Failover.File, `
mode: disabled
state_provider: stateboard
stateboard_params:
  uri: yuriy
  password: stroganov-bmstu
fencing_enabled: false
fencing_timeout: 380`)

	defer os.Remove(ctx.Failover.File)
	assert.Equal(nil, err)

	_, err = getFailoverOptsFromFile(&ctx)
	assert.Equal(nil, err)
}

func createYmlFileWithContent(fileName string, content string) error {
	failoverFile, err := os.Create(fileName)
	if err != nil {
		return nil
	}

	_, err = failoverFile.WriteString(content)
	return err
}
