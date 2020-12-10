package replicasets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetReplicasetSummary(t *testing.T) {
	assert := assert.New(t)

	var topologyReplicaset *TopologyReplicaset
	var summary string
	var expSummary string

	vshardGroup := "hot"
	weight := 123.4
	allRWTrue := true
	allRWFalse := false

	// simplest replicaset
	topologyReplicaset = &TopologyReplicaset{
		UUID:   "rpl-uuid",
		Alias:  "rpl-alias",
		Status: "healthy",
		Roles:  []string{},
		Instances: TopologyInstances{
			&TopologyInstance{
				Alias: "instance-1",
				UUID:  "uuid-1",
				URI:   "uri-1",
			},
		},
	}

	summary = getTopologyReplicasetSummary(topologyReplicaset)
	expSummary = `• rpl-alias
  No roles
    • instance-1 uri-1`

	assert.Equal(expSummary, summary)

	// replicaset w/ roles
	topologyReplicaset = &TopologyReplicaset{
		UUID:   "rpl-uuid",
		Alias:  "rpl-alias",
		Status: "healthy",
		Roles:  []string{"role-1", "role-2", "role-3"},
		Instances: TopologyInstances{
			&TopologyInstance{
				Alias: "instance-1",
				UUID:  "uuid-1",
				URI:   "uri-1",
			},
		},
	}

	summary = getTopologyReplicasetSummary(topologyReplicaset)
	expSummary = `• rpl-alias
  Role: role-1 | role-2 | role-3
    • instance-1 uri-1`

	assert.Equal(expSummary, summary)

	// replicaset w/ additional params
	topologyReplicaset = &TopologyReplicaset{
		UUID:        "rpl-uuid",
		Alias:       "rpl-alias",
		Status:      "healthy",
		Roles:       []string{},
		VshardGroup: &vshardGroup,
		Weight:      &weight,
		AllRW:       &allRWTrue,
		Instances: TopologyInstances{
			&TopologyInstance{
				Alias: "instance-1",
				UUID:  "uuid-1",
				URI:   "uri-1",
			},
		},
	}

	summary = getTopologyReplicasetSummary(topologyReplicaset)
	expSummary = `• rpl-alias                       hot | 123.4 | ALL RW
  No roles
    • instance-1 uri-1`

	assert.Equal(expSummary, summary)

	// replicaset w/ all rw false
	topologyReplicaset = &TopologyReplicaset{
		UUID:        "rpl-uuid",
		Alias:       "rpl-alias",
		Status:      "healthy",
		Roles:       []string{},
		VshardGroup: &vshardGroup,
		Weight:      &weight,
		AllRW:       &allRWFalse,
		Instances: TopologyInstances{
			&TopologyInstance{
				Alias: "instance-1",
				UUID:  "uuid-1",
				URI:   "uri-1",
			},
		},
	}

	summary = getTopologyReplicasetSummary(topologyReplicaset)
	expSummary = `• rpl-alias                       hot | 123.4
  No roles
    • instance-1 uri-1`

	assert.Equal(expSummary, summary)

	// replicaset w/ leader
	topologyReplicaset = &TopologyReplicaset{
		UUID:       "rpl-uuid",
		Alias:      "rpl-alias",
		Status:     "healthy",
		Roles:      []string{},
		LeaderUUID: "uuid-2",
		Instances: TopologyInstances{
			&TopologyInstance{
				Alias: "instance-1",
				UUID:  "uuid-1",
				URI:   "uri-1",
			},
			&TopologyInstance{
				Alias: "instance-2",
				UUID:  "uuid-2",
				URI:   "uri-2",
			},
			&TopologyInstance{
				Alias: "instance-3",
				UUID:  "uuid-3",
				URI:   "uri-3",
			},
		},
	}

	summary = getTopologyReplicasetSummary(topologyReplicaset)
	expSummary = `• rpl-alias
  No roles
    • instance-1 uri-1
    ★ instance-2 uri-2
    • instance-3 uri-3`

	assert.Equal(expSummary, summary)

	// instances w/ zones
	topologyReplicaset = &TopologyReplicaset{
		UUID:   "rpl-uuid",
		Alias:  "rpl-alias",
		Status: "healthy",
		Roles:  []string{},
		Instances: TopologyInstances{
			&TopologyInstance{
				Alias: "instance-1",
				UUID:  "uuid-1",
				URI:   "uri-1",
				Zone:  "msk",
			},
			&TopologyInstance{
				Alias: "instance-2",
				UUID:  "uuid-2",
				URI:   "uri-2",
				Zone:  "spb",
			},
			&TopologyInstance{
				Alias: "instance-3",
				UUID:  "uuid-3",
				URI:   "uri-3",
			},
		},
	}

	summary = getTopologyReplicasetSummary(topologyReplicaset)
	expSummary = `• rpl-alias
  No roles
    • instance-1 uri-1                         msk
    • instance-2 uri-2                         spb
    • instance-3 uri-3`

	assert.Equal(expSummary, summary)

	// replicaset w/ everything
	topologyReplicaset = &TopologyReplicaset{
		UUID:        "rpl-uuid",
		Alias:       "rpl-alias",
		Status:      "healthy",
		Roles:       []string{"role-1", "role-2", "role-3"},
		VshardGroup: &vshardGroup,
		Weight:      &weight,
		AllRW:       &allRWTrue,
		LeaderUUID:  "uuid-2",
		Instances: TopologyInstances{
			&TopologyInstance{
				Alias: "instance-1",
				UUID:  "uuid-1",
				URI:   "uri-1",
				Zone:  "msk",
			},
			&TopologyInstance{
				Alias: "instance-2",
				UUID:  "uuid-2",
				URI:   "uri-2",
				Zone:  "spb",
			},
			&TopologyInstance{
				Alias: "instance-3",
				UUID:  "uuid-3",
				URI:   "uri-3",
			},
		},
	}

	summary = getTopologyReplicasetSummary(topologyReplicaset)
	expSummary = `• rpl-alias                       hot | 123.4 | ALL RW
  Role: role-1 | role-2 | role-3
    • instance-1 uri-1                         msk
    ★ instance-2 uri-2                         spb
    • instance-3 uri-3`

	assert.Equal(expSummary, summary)
}
