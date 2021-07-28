package failover

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarantool/cartridge-cli/cli/context"
	"gopkg.in/yaml.v2"
)

func TestValidateFailoverYMLFile(t *testing.T) {
	assert := assert.New(t)

	okConfigurations := []map[string]interface{}{
		// Disabled mode
		{
			"mode": "disabled",
		},
		// Eventual mode
		{
			"mode":             "eventual",
			"failover_timeout": 1,
			"fencing_enabled":  true,
			"fencing_timeout":  88,
			"fencing_pause":    4,
		},
		// Stateful mode + stateboard provider
		{
			"mode":           "stateful",
			"state_provider": "stateboard",
			"stateboard_params": map[string]interface{}{
				"uri":      "yuriy",
				"password": "stroganov-bmstu",
			},
			"fencing_enabled": false,
			"fencing_timeout": 380,
		},
		// Stateful mode + etcd2 provider
		{
			"mode":           "stateful",
			"state_provider": "etcd2",
			"etcd2_params": map[string]interface{}{
				"prefix":     "xiferp",
				"lock_delay": 120,
				"endpoints":  []string{"http://localhost:2379", "http://localhost:4001"},
				"password":   "superpass",
				"username":   "superuser",
			},
		},
		// Stateful mode + stateboard provider and etcd2 parameters
		{
			"mode":           "stateful",
			"state_provider": "stateboard",
			"stateboard_params": map[string]interface{}{
				"uri":      "uri",
				"password": "pass",
			},
			"etcd2_params": map[string]interface{}{
				"uri":      "uri",
				"password": "pass",
			},
		},
		// Stateful mode + etcd2 provider and no etcd2 parameters
		{
			"mode":           "stateful",
			"state_provider": "etcd2",
		},
		// Stateful mode + etcd2 provider and stateboard parameters
		{
			"mode": "stateful",
			"stateboard_params": map[string]interface{}{
				"uri":      "uri",
				"password": "pass",
			},
		},
		// Disabled mode + stateboard provider and stateboard paramters
		{
			"mode":           "disabled",
			"state_provider": "stateboard",
			"stateboard_params": map[string]interface{}{
				"uri":      "yuriy",
				"password": "stroganov-bmstu",
			},
			"fencing_enabled": false,
			"fencing_timeout": 380,
		},
	}

	ctx := context.Ctx{}
	ctx.Failover.File = "failover_validate_test"

	for _, conf := range okConfigurations {
		err := createYmlFileWithContent(ctx.Failover.File, conf)
		defer os.Remove(ctx.Failover.File)
		assert.Equal(nil, err)

		_, err = getFailoverOptsFromFile(&ctx)
		assert.Equal(nil, err)
	}
}

func createYmlFileWithContent(fileName string, content map[string]interface{}) error {
	failoverFile, err := os.Create(fileName)
	if err != nil {
		return nil
	}

	yaml.NewEncoder(failoverFile).Encode(content)
	return nil
}
