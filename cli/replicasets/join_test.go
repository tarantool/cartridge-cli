package replicasets

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetJoinInstancesEditReplicasetsOpts(t *testing.T) {
	assert := assert.New(t)

	var err error
	var opts *EditReplicasetOpts
	var serializedOpts string
	var joinInstancesNames []string

	topologyReplicasets := &TopologyReplicasets{
		"rpl-1-uuid": &TopologyReplicaset{
			UUID:   "rpl-1-uuid",
			Alias:  "rpl-1-alias",
			Status: "healthy",
			Roles:  []string{"some-role", "other-role"},
			Instances: TopologyInstances{
				&TopologyInstance{
					Alias: "instance-1",
					UUID:  "uuid-1",
				},
			},
		},
		"rpl-2-uuid": &TopologyReplicaset{
			UUID:   "rpl-2-uuid",
			Alias:  "rpl-2-alias",
			Status: "healthy",
			Roles:  []string{"some-role", "other-role"},
			Instances: TopologyInstances{
				&TopologyInstance{
					Alias: "instance-2",
					UUID:  "uuid-2",
				},
			},
		},
	}

	instancesConf := &InstancesConf{
		"instance-1": &InstanceConf{URI: "uri-1"},
		"instance-2": &InstanceConf{URI: "uri-2"},
		"instance-3": &InstanceConf{URI: "uri-3"},
		"instance-4": &InstanceConf{URI: "uri-4"},
	}

	// create a new replicaset

	joinInstancesNames = []string{"instance-3", "instance-4"}
	opts, err = getJoinInstancesEditReplicasetsOpts("rpl-3-alias", joinInstancesNames, topologyReplicasets, instancesConf)
	assert.Nil(err)

	// create new replicaset and specify it's alias
	assert.Equal("rpl-3-alias", opts.ReplicasetAlias)
	assert.Equal([]string{"uri-3", "uri-4"}, opts.JoinInstancesURIs)

	serializedOpts = serializeEditReplicasetOpts(opts)
	assert.Equal(
		"alias = 'rpl-3-alias', join_servers = { { uri = 'uri-3' }, { uri = 'uri-4' } }",
		serializedOpts,
	)

	// join to the existent one

	joinInstancesNames = []string{"instance-3", "instance-4"}
	opts, err = getJoinInstancesEditReplicasetsOpts("rpl-2-alias", joinInstancesNames, topologyReplicasets, instancesConf)
	assert.Nil(err)

	// join to the existent replicaset by uuid
	assert.Equal("rpl-2-uuid", opts.ReplicasetUUID)
	assert.Equal([]string{"uri-3", "uri-4"}, opts.JoinInstancesURIs)

	serializedOpts = serializeEditReplicasetOpts(opts)
	assert.Equal(
		"uuid = 'rpl-2-uuid', join_servers = { { uri = 'uri-3' }, { uri = 'uri-4' } }",
		serializedOpts,
	)

	// unknown instance name specified

	joinInstancesNames = []string{"instance-3", "instance-4", "unknown-instance"}
	opts, err = getJoinInstancesEditReplicasetsOpts("rpl-2-alias", joinInstancesNames, topologyReplicasets, instancesConf)
	assert.True(strings.Contains(err.Error(), `Configuration for instance unknown-instance hasn't found in instances.yml`), err.Error())
}
