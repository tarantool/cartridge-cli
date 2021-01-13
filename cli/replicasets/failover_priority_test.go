package replicasets

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFailoverPriorityEditReplicasetOpts(t *testing.T) {
	assert := assert.New(t)

	var err error
	var instanceNames []string
	var opts *EditReplicasetOpts

	topologyReplicaset := &TopologyReplicaset{
		UUID:   "replicaset-uuid",
		Alias:  "replicaset-alias",
		Status: "healthy",
		Roles:  []string{"some-role", "other-role"},
		Instances: TopologyInstances{
			&TopologyInstance{
				Alias: "instance-1",
				UUID:  "uuid-1",
			},
			&TopologyInstance{
				Alias: "instance-2",
				UUID:  "uuid-2",
			},
			&TopologyInstance{
				Alias: "instance-3",
				UUID:  "uuid-3",
			},
		},
	}

	// everything is OK
	instanceNames = []string{"instance-3", "instance-2"}

	opts, err = getSetFailoverPriorityEditReplicasetOpts(instanceNames, topologyReplicaset)
	assert.Nil(err)
	assert.Equal(topologyReplicaset.UUID, opts.ReplicasetUUID)
	assert.Equal([]string{"uuid-3", "uuid-2"}, opts.FailoverPriorityUUIDs)

	// specified unknown instance alias
	instanceNames = []string{"unknown-instance", "instance-3", "instance-2"}

	opts, err = getSetFailoverPriorityEditReplicasetOpts(instanceNames, topologyReplicaset)
	assert.True(strings.Contains(err.Error(), `Instance unknown-instance not found in replica set`), err.Error())
}
