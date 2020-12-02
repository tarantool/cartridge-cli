package replicasets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpelEditInstancesOpts(t *testing.T) {
	assert := assert.New(t)

	var instancesToExpelUUIDs []string
	var err error
	var opts *EditInstancesListOpts
	var serializedOpts string

	// no uuids are specified
	instancesToExpelUUIDs = []string{}

	opts, err = getExpelInstancesEditInstancesOpts(instancesToExpelUUIDs)
	assert.Nil(err)
	assert.Len(*opts, 0)

	serializedOpts = serializeEditInstancesListOpts(opts)
	assert.Equal("", serializedOpts)

	// uuids are specified
	instancesToExpelUUIDs = []string{
		"uuid-1", "uuid-2", "uuid-3",
	}

	opts, err = getExpelInstancesEditInstancesOpts(instancesToExpelUUIDs)
	assert.Nil(err)
	assert.Len(*opts, len(instancesToExpelUUIDs))

	for i, uuid := range instancesToExpelUUIDs {
		opt := (*opts)[i]
		expOpt := EditInstanceOpts{
			InstanceUUID: uuid,
			Expelled:     true,
		}
		assert.Equal(expOpt, *opt)
	}

	serializedOpts = serializeEditInstancesListOpts(opts)
	assert.Equal(
		"{ uuid = 'uuid-1', expelled = true }, { uuid = 'uuid-2', expelled = true }, { uuid = 'uuid-3', expelled = true }",
		serializedOpts,
	)
}
