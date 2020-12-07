package replicasets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSetWeightEditReplicasetOpts(t *testing.T) {
	assert := assert.New(t)

	var err error
	var oldWeight float64
	var newWeight float64
	var opts *EditReplicasetOpts
	var serializedOpts string

	oldWeight = 1

	topologyReplicaset := &TopologyReplicaset{
		UUID:   "replicaset-uuid",
		Alias:  "replicaset-alias",
		Status: "healthy",
		Roles:  []string{"some-role", "other-role"},
		Weight: &oldWeight,
		Instances: TopologyInstances{
			&TopologyInstance{
				Alias: "instance-1",
				UUID:  "uuid-1",
			},
		},
	}

	// int
	newWeight = 111

	opts, err = getSetWeightEditReplicasetOpts(newWeight, topologyReplicaset)
	assert.Nil(err)
	assert.Equal(topologyReplicaset.UUID, opts.ReplicasetUUID)
	assert.Equal(newWeight, *opts.Weight)

	serializedOpts = serializeEditReplicasetOpts(opts)
	assert.Equal(
		"uuid = 'replicaset-uuid', weight = 111",
		serializedOpts,
	)

	// float
	newWeight = 111.123

	opts, err = getSetWeightEditReplicasetOpts(newWeight, topologyReplicaset)
	assert.Nil(err)
	assert.Equal(topologyReplicaset.UUID, opts.ReplicasetUUID)
	assert.Equal(newWeight, *opts.Weight)

	serializedOpts = serializeEditReplicasetOpts(opts)
	assert.Equal(
		"uuid = 'replicaset-uuid', weight = 111.123",
		serializedOpts,
	)
}
