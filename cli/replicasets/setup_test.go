package replicasets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCreateReplicasetEditReplicasetsOpts(t *testing.T) {
	assert := assert.New(t)

	var err error
	var replicasetConf *ReplicasetConf
	var opts *EditReplicasetOpts
	var serializedOpts string

	var allRW bool
	var weight float64
	var vshardGroup string

	instancesConf := &InstancesConf{
		"instance-1": &InstanceConf{URI: "uri-1"},
		"instance-2": &InstanceConf{URI: "uri-2"},
		"instance-3": &InstanceConf{URI: "uri-3"},
		"instance-4": &InstanceConf{URI: "uri-4"},
	}

	// create replicaset w/o all_rw, weight and vshard_group specified
	replicasetConf = &ReplicasetConf{
		Alias:       "rpl-alias",
		Roles:       []string{"some-role", "other-role"},
		AllRW:       nil,
		Weight:      nil,
		VshardGroup: nil,
		InstanceNames: []string{
			"instance-1", "instance-2", "instance-4",
		},
	}

	opts, err = getCreateReplicasetEditReplicasetsOpts(replicasetConf, instancesConf)
	assert.Nil(err)
	assert.Equal(replicasetConf.Alias, opts.ReplicasetAlias)
	assert.Equal(replicasetConf.Roles, opts.Roles)
	assert.Nil(opts.AllRW)
	assert.Nil(opts.Weight)
	assert.Nil(opts.VshardGroup)
	assert.Nil(opts.FailoverPriorityUUIDs)
	assert.Equal([]string{"uri-1", "uri-2", "uri-4"}, opts.JoinInstancesURIs)

	serializedOpts, err = getEditReplicasetOptsString(opts)
	assert.Nil(err)
	assert.Equal(
		"{ alias = 'rpl-alias', roles = { 'some-role', 'other-role' }, "+
			"join_servers = { { uri = 'uri-1' }, { uri = 'uri-2' }, { uri = 'uri-4' } } }",
		serializedOpts,
	)

	// create replicaset w all_rw, weight and vshard_group specified
	allRW = true
	weight = 123.4
	vshardGroup = "hot"

	replicasetConf = &ReplicasetConf{
		Alias:       "rpl-alias",
		Roles:       []string{"some-role", "other-role"},
		AllRW:       &allRW,
		Weight:      &weight,
		VshardGroup: &vshardGroup,
		InstanceNames: []string{
			"instance-1", "instance-2", "instance-4",
		},
	}

	opts, err = getCreateReplicasetEditReplicasetsOpts(replicasetConf, instancesConf)
	assert.Nil(err)
	assert.Equal(replicasetConf.Alias, opts.ReplicasetAlias)
	assert.Equal(replicasetConf.Roles, opts.Roles)
	assert.True(*opts.AllRW)
	assert.Equal(weight, *opts.Weight)
	assert.Equal(vshardGroup, *opts.VshardGroup)
	assert.Nil(opts.FailoverPriorityUUIDs)
	assert.Equal([]string{"uri-1", "uri-2", "uri-4"}, opts.JoinInstancesURIs)

	serializedOpts, err = getEditReplicasetOptsString(opts)
	assert.Nil(err)
	assert.Equal(
		"{ alias = 'rpl-alias', roles = { 'some-role', 'other-role' }, "+
			"all_rw = true, weight = 123.4, vshard_group = 'hot', "+
			"join_servers = { { uri = 'uri-1' }, { uri = 'uri-2' }, { uri = 'uri-4' } } }",
		serializedOpts,
	)
}

func TestGetUpdateReplicasetEditReplicasetsOpts(t *testing.T) {
	assert := assert.New(t)

	var err error
	var replicasetConf *ReplicasetConf
	var opts *EditReplicasetOpts
	var serializedOpts string

	var oldAllRW bool
	var allRW bool
	var oldWeight float64
	var weight float64
	var oldVshardGroup string
	var vshardGroup string

	instancesConf := &InstancesConf{
		"instance-1": &InstanceConf{URI: "uri-1"},
		"instance-2": &InstanceConf{URI: "uri-2"},
		"instance-3": &InstanceConf{URI: "uri-3"},
		"instance-4": &InstanceConf{URI: "uri-4"},
	}

	oldAllRW = false
	oldWeight = 987.65
	oldVshardGroup = "cold"

	topologyReplicaset := &TopologyReplicaset{
		UUID:        "rpl-uuid",
		Alias:       "rpl-alias",
		Status:      "healthy",
		Roles:       []string{"some-role", "other-role"},
		AllRW:       &oldAllRW,
		Weight:      &oldWeight,
		VshardGroup: &oldVshardGroup,
		Instances: TopologyInstances{
			&TopologyInstance{
				Alias: "instance-1",
				UUID:  "uuid-1",
			},
			&TopologyInstance{
				Alias: "instance-3",
				UUID:  "uuid-3",
			},
		},
	}

	// update replicaset w/o all_rw, weight and vshard_group specified
	replicasetConf = &ReplicasetConf{
		Alias:       "rpl-alias",
		Roles:       []string{"some-other-role", "one-more-other-role"},
		AllRW:       nil,
		Weight:      nil,
		VshardGroup: nil,
		InstanceNames: []string{
			"instance-1", "instance-2", "instance-4",
		},
	}

	opts, err = getUpdateReplicasetEditReplicasetsOpts(topologyReplicaset, replicasetConf, instancesConf)
	assert.Nil(err)
	assert.Equal(topologyReplicaset.UUID, opts.ReplicasetUUID)
	assert.Equal(replicasetConf.Roles, opts.Roles)
	assert.Nil(opts.AllRW)
	assert.Nil(opts.Weight)
	assert.Nil(opts.VshardGroup)
	assert.Nil(opts.FailoverPriorityUUIDs)
	assert.Equal([]string{"uri-2", "uri-4"}, opts.JoinInstancesURIs)

	serializedOpts, err = getEditReplicasetOptsString(opts)
	assert.Nil(err)
	assert.Equal(
		"{ uuid = 'rpl-uuid', roles = { 'some-other-role', 'one-more-other-role' }, "+
			"join_servers = { { uri = 'uri-2' }, { uri = 'uri-4' } } }",
		serializedOpts,
	)

	// update replicaset w/o all_rw, weight and vshard_group specified
	allRW = true
	weight = 123.4
	vshardGroup = "hot"

	replicasetConf = &ReplicasetConf{
		Alias:       "rpl-alias",
		Roles:       []string{"some-other-role", "one-more-other-role"},
		AllRW:       &allRW,
		Weight:      &weight,
		VshardGroup: &vshardGroup,
		InstanceNames: []string{
			"instance-1", "instance-2", "instance-4",
		},
	}

	opts, err = getUpdateReplicasetEditReplicasetsOpts(topologyReplicaset, replicasetConf, instancesConf)
	assert.Nil(err)
	assert.Equal(topologyReplicaset.UUID, opts.ReplicasetUUID)
	assert.Equal(replicasetConf.Roles, opts.Roles)
	assert.True(*opts.AllRW)
	assert.Equal(weight, *opts.Weight)
	assert.Equal(vshardGroup, *opts.VshardGroup)
	assert.Nil(opts.FailoverPriorityUUIDs)
	assert.Equal([]string{"uri-2", "uri-4"}, opts.JoinInstancesURIs)

	serializedOpts, err = getEditReplicasetOptsString(opts)
	assert.Nil(err)
	assert.Equal(
		"{ uuid = 'rpl-uuid', roles = { 'some-other-role', 'one-more-other-role' }, "+
			"all_rw = true, weight = 123.4, vshard_group = 'hot', "+
			"join_servers = { { uri = 'uri-2' }, { uri = 'uri-4' } } }",
		serializedOpts,
	)
}
