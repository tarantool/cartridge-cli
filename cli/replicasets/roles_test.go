package replicasets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUpdateRolesEditReplicasetsOpts(t *testing.T) {
	assert := assert.New(t)

	var err error
	var specifiedRoles []string
	var vshardGroup string
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
		},
	}

	// add roles, vshard group is specified

	specifiedRoles = []string{"some-new-role", "some-role"}
	vshardGroup = "some-group"

	opts, err = getUpdateRolesEditReplicasetsOpts(addRolesToList, specifiedRoles, vshardGroup, topologyReplicaset)
	assert.Nil(err)
	assert.Equal("replicaset-uuid", opts.ReplicasetUUID)
	assert.Equal([]string{"other-role", "some-new-role", "some-role"}, opts.Roles)
	assert.Equal(vshardGroup, *opts.VshardGroup)

	specifiedRoles = []string{"some-not-added-role", "some-role"}

	opts, err = getUpdateRolesEditReplicasetsOpts(removeRolesFromList, specifiedRoles, "", topologyReplicaset)
	assert.Nil(err)
	assert.Equal("replicaset-uuid", opts.ReplicasetUUID)
	assert.Equal([]string{"other-role"}, opts.Roles)
}
