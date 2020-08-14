package repair

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func writeTopologyConfig(workDir string, content string) {
	configPath := filepath.Join(workDir, "config", "topology.yml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0777); err != nil {
		panic(fmt.Errorf("Failed to create clusterwide config directory: %s", err))
	}

	if err := ioutil.WriteFile(configPath, []byte(content), 0644); err != nil {
		panic(fmt.Errorf("Failed to write clusterwide config: %s", err))
	}
}

func writeOneFileConfig(workDir string, content string) {
	configPath := filepath.Join(workDir, "config.yml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0777); err != nil {
		panic(fmt.Errorf("Failed to create clusterwide config directory: %s", err))
	}

	if err := ioutil.WriteFile(configPath, []byte(content), 0644); err != nil {
		panic(fmt.Errorf("Failed to write clusterwide config: %s", err))
	}
}

func TestGetTopologyConf(t *testing.T) {
	assert := assert.New(t)

	var err error
	var topologyConf *TopologyConfType
	var instanceConf *InstanceConfType
	var replicasetConf *ReplicasetConfType

	// create tmp working directory
	workDir, err := ioutil.TempDir("", "work-dir")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(workDir)

	writeTopologyConfig(workDir, `---
failover: false
replicasets:
  rpl-1:
    alias: replicaset-1
    all_rw: false
    master:
    - srv-1
    roles:
      app.roles.custom: true
      failover-coordinator: true
      vshard-router: true
    weight: 0
  rpl-2:
    alias: unnamed
    all_rw: false
    master:
    - srv-2
    - srv-disabled
    roles:
      vshard-storage: true
    vshard_group: default
    weight: 1
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: localhost:3301
  srv-2:
    disabled: false
    replicaset_uuid: rpl-2
    uri: localhost:3302
  srv-not-in-master:
    disabled: false
    uri: localhost:3303
    replicaset_uuid: rpl-1
  srv-disabled:
    disabled: true
    uri: localhost:3304
    replicaset_uuid: rpl-2
  srv-expelled: expelled
`)

	topologyConf, err = getTopologyConf(workDir)
	assert.Nil(err)
	assert.Equal(filepath.Join(workDir, "config", "topology.yml"), topologyConf.Path)
	assert.False(topologyConf.ConfIsOneFile)

	// instances
	assert.Equal(5, len(topologyConf.Instances))

	assert.Contains(topologyConf.Instances, "srv-1")
	instanceConf, _ = topologyConf.Instances["srv-1"]
	assert.Equal(instanceConf.AdvertiseURI, "localhost:3301")
	assert.Equal(instanceConf.ReplicasetUUID, "rpl-1")
	assert.False(instanceConf.IsDisabled)
	assert.False(instanceConf.IsExpelled)

	assert.Contains(topologyConf.Instances, "srv-2")
	instanceConf, _ = topologyConf.Instances["srv-2"]
	assert.Equal(instanceConf.AdvertiseURI, "localhost:3302")
	assert.Equal(instanceConf.ReplicasetUUID, "rpl-2")
	assert.False(instanceConf.IsDisabled)
	assert.False(instanceConf.IsExpelled)

	assert.Contains(topologyConf.Instances, "srv-not-in-master")
	instanceConf, _ = topologyConf.Instances["srv-not-in-master"]
	assert.Equal(instanceConf.AdvertiseURI, "localhost:3303")
	assert.Equal(instanceConf.ReplicasetUUID, "rpl-1")
	assert.False(instanceConf.IsDisabled)
	assert.False(instanceConf.IsExpelled)

	assert.Contains(topologyConf.Instances, "srv-disabled")
	instanceConf, _ = topologyConf.Instances["srv-disabled"]
	assert.Equal(instanceConf.AdvertiseURI, "localhost:3304")
	assert.Equal(instanceConf.ReplicasetUUID, "rpl-2")
	assert.True(instanceConf.IsDisabled)
	assert.False(instanceConf.IsExpelled)

	assert.Contains(topologyConf.Instances, "srv-expelled")
	instanceConf, _ = topologyConf.Instances["srv-expelled"]
	assert.Equal(instanceConf.AdvertiseURI, "")
	assert.Equal(instanceConf.ReplicasetUUID, "")
	assert.False(instanceConf.IsDisabled)
	assert.True(instanceConf.IsExpelled)

	// replicasets
	assert.Equal(2, len(topologyConf.Replicasets))

	assert.Contains(topologyConf.Replicasets, "rpl-1")
	replicasetConf, _ = topologyConf.Replicasets["rpl-1"]
	assert.False(replicasetConf.LeadersIsString)
	assert.Equal("replicaset-1", replicasetConf.Alias)
	assert.ElementsMatch([]string{"srv-1", "srv-not-in-master"}, replicasetConf.Instances)
	assert.Equal([]string{"srv-1"}, replicasetConf.Leaders)
	assert.ElementsMatch([]string{"app.roles.custom", "failover-coordinator", "vshard-router"}, replicasetConf.Roles)

	assert.Contains(topologyConf.Replicasets, "rpl-2")
	replicasetConf, _ = topologyConf.Replicasets["rpl-2"]
	assert.False(replicasetConf.LeadersIsString)
	assert.Equal("", replicasetConf.Alias)
	assert.ElementsMatch([]string{"srv-2", "srv-disabled"}, replicasetConf.Instances)
	assert.Equal([]string{"srv-2", "srv-disabled"}, replicasetConf.Leaders)
	assert.ElementsMatch([]string{"vshard-storage"}, replicasetConf.Roles)
}

func TestGetTopologyConfOneFile(t *testing.T) {
	assert := assert.New(t)

	var err error
	var topologyConf *TopologyConfType
	var instanceConf *InstanceConfType
	var replicasetConf *ReplicasetConfType

	// create tmp working directory
	workDir, err := ioutil.TempDir("", "work-dir")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(workDir)

	writeOneFileConfig(workDir, `---
auth: {}
topology:
  failover: false
  replicasets:
    rpl-1:
      alias: replicaset-1
      all_rw: false
      master:
      - srv-1
      roles:
        app.roles.custom: true
        failover-coordinator: true
        vshard-router: true
      weight: 0
    rpl-2:
      alias: unnamed
      all_rw: false
      master:
      - srv-2
      - srv-disabled
      roles:
        vshard-storage: true
      vshard_group: default
      weight: 1
  servers:
    srv-1:
      disabled: false
      replicaset_uuid: rpl-1
      uri: localhost:3301
    srv-2:
      disabled: false
      replicaset_uuid: rpl-2
      uri: localhost:3302
    srv-not-in-master:
      disabled: false
      uri: localhost:3303
      replicaset_uuid: rpl-1
    srv-disabled:
      disabled: true
      uri: localhost:3304
      replicaset_uuid: rpl-2
    srv-expelled: expelled
`)

	topologyConf, err = getTopologyConf(workDir)
	assert.Nil(err)
	assert.Equal(filepath.Join(workDir, "config.yml"), topologyConf.Path)
	assert.True(topologyConf.ConfIsOneFile)

	// instances
	assert.Equal(5, len(topologyConf.Instances))

	assert.Contains(topologyConf.Instances, "srv-1")
	instanceConf, _ = topologyConf.Instances["srv-1"]
	assert.Equal(instanceConf.AdvertiseURI, "localhost:3301")
	assert.Equal(instanceConf.ReplicasetUUID, "rpl-1")
	assert.False(instanceConf.IsDisabled)
	assert.False(instanceConf.IsExpelled)

	assert.Contains(topologyConf.Instances, "srv-2")
	instanceConf, _ = topologyConf.Instances["srv-2"]
	assert.Equal(instanceConf.AdvertiseURI, "localhost:3302")
	assert.Equal(instanceConf.ReplicasetUUID, "rpl-2")
	assert.False(instanceConf.IsDisabled)
	assert.False(instanceConf.IsExpelled)

	assert.Contains(topologyConf.Instances, "srv-not-in-master")
	instanceConf, _ = topologyConf.Instances["srv-not-in-master"]
	assert.Equal(instanceConf.AdvertiseURI, "localhost:3303")
	assert.Equal(instanceConf.ReplicasetUUID, "rpl-1")
	assert.False(instanceConf.IsDisabled)
	assert.False(instanceConf.IsExpelled)

	assert.Contains(topologyConf.Instances, "srv-disabled")
	instanceConf, _ = topologyConf.Instances["srv-disabled"]
	assert.Equal(instanceConf.AdvertiseURI, "localhost:3304")
	assert.Equal(instanceConf.ReplicasetUUID, "rpl-2")
	assert.True(instanceConf.IsDisabled)
	assert.False(instanceConf.IsExpelled)

	assert.Contains(topologyConf.Instances, "srv-expelled")
	instanceConf, _ = topologyConf.Instances["srv-expelled"]
	assert.Equal(instanceConf.AdvertiseURI, "")
	assert.Equal(instanceConf.ReplicasetUUID, "")
	assert.False(instanceConf.IsDisabled)
	assert.True(instanceConf.IsExpelled)

	// replicasets
	assert.Equal(2, len(topologyConf.Replicasets))

	assert.Contains(topologyConf.Replicasets, "rpl-1")
	replicasetConf, _ = topologyConf.Replicasets["rpl-1"]
	assert.Equal("replicaset-1", replicasetConf.Alias)
	assert.ElementsMatch([]string{"srv-1", "srv-not-in-master"}, replicasetConf.Instances)
	assert.Equal([]string{"srv-1"}, replicasetConf.Leaders)
	assert.ElementsMatch([]string{"app.roles.custom", "failover-coordinator", "vshard-router"}, replicasetConf.Roles)

	assert.Contains(topologyConf.Replicasets, "rpl-2")
	replicasetConf, _ = topologyConf.Replicasets["rpl-2"]
	assert.Equal("", replicasetConf.Alias)
	assert.ElementsMatch([]string{"srv-2", "srv-disabled"}, replicasetConf.Instances)
	assert.Equal([]string{"srv-2", "srv-disabled"}, replicasetConf.Leaders)
	assert.ElementsMatch([]string{"vshard-storage"}, replicasetConf.Roles)
}

func TestSetInstanceURI(t *testing.T) {
	assert := assert.New(t)

	var err error
	var topologyConf *TopologyConfType
	var newContent []byte

	// create tmp working directory
	workDir, err := ioutil.TempDir("", "work-dir")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(workDir)

	writeTopologyConfig(workDir, `---
failover: false
replicasets: {}
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: localhost:3301
  srv-expelled: expelled
`)

	topologyConf, err = getTopologyConf(workDir)
	assert.Nil(err)

	err = topologyConf.SetInstanceURI("srv-1", "localhost:3311")
	assert.Nil(err)
	fmt.Printf("topologyConf.Instances: %#v\n", topologyConf.Instances)
	assert.Equal("localhost:3311", topologyConf.Instances["srv-1"].AdvertiseURI)

	newContent, err = topologyConf.MarshalContent()
	assert.Nil(err)

	expContent := `failover: false
replicasets: {}
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: localhost:3311
  srv-expelled: expelled
`

	assert.Equal(expContent, string(newContent))
}

func TestRemoveInstance(t *testing.T) {
	assert := assert.New(t)

	var err error
	var topologyConf *TopologyConfType
	var newContent []byte

	// create tmp working directory
	workDir, err := ioutil.TempDir("", "work-dir")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(workDir)

	writeTopologyConfig(workDir, `---
failover: false
replicasets:
  rpl-1:
    alias: replicaset-1
    all_rw: false
    master:
    - srv-1
    roles:
      app.roles.custom: true
      failover-coordinator: true
      vshard-router: true
    weight: 0
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: localhost:3301
  srv-expelled: expelled
`)

	topologyConf, err = getTopologyConf(workDir)
	assert.Nil(err)

	err = topologyConf.RemoveInstance("srv-non-existant")
	assert.EqualError(err, "Instance srv-non-existant isn't found in cluster")

	err = topologyConf.RemoveInstance("srv-1")
	assert.Nil(err)
	assert.NotContains(topologyConf.Instances, "srv-1")

	newContent, err = topologyConf.MarshalContent()
	assert.Nil(err)

	expContent := `failover: false
replicasets:
  rpl-1:
    alias: replicaset-1
    all_rw: false
    master:
    - srv-1
    roles:
      app.roles.custom: true
      failover-coordinator: true
      vshard-router: true
    weight: 0
servers:
  srv-expelled: expelled
`

	assert.Equal(expContent, string(newContent))
}

func RemoveReplicaset(t *testing.T) {
	assert := assert.New(t)

	var err error
	var topologyConf *TopologyConfType
	var newContent []byte

	// create tmp working directory
	workDir, err := ioutil.TempDir("", "work-dir")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(workDir)

	writeTopologyConfig(workDir, `---
failover: false
replicasets:
  rpl-1:
    alias: replicaset-1
    all_rw: false
    master:
    - srv-1
    roles:
      app.roles.custom: true
      failover-coordinator: true
      vshard-router: true
    weight: 0
  rpl-2:
    alias: unnamed
    all_rw: false
    master:
    - srv-2
    - srv-disabled
    roles:
      vshard-storage: true
    vshard_group: default
    weight: 1
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: localhost:3301
  srv-expelled: expelled
`)

	topologyConf, err = getTopologyConf(workDir)
	assert.Nil(err)

	err = topologyConf.RemoveReplicaset("rpl-non-existent")
	assert.EqualError(err, "Replicaset rpl-non-existant isn't found in cluster")

	err = topologyConf.RemoveReplicaset("rpl-2")
	assert.Nil(err)
	assert.NotContains(topologyConf.Replicasets, "rpl-2")

	newContent, err = topologyConf.MarshalContent()
	assert.Nil(err)

	expContent := `failover: false
replicasets:
  rpl-1:
    alias: replicaset-1
    all_rw: false
    master:
    - srv-1
    roles:
      app.roles.custom: true
      failover-coordinator: true
      vshard-router: true
    weight: 0
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: localhost:3301
  srv-expelled: expelled
`

	assert.Equal(expContent, string(newContent))
}

func TestSetInstances(t *testing.T) {
	assert := assert.New(t)

	var err error
	var topologyConf *TopologyConfType
	var newContent []byte

	// create tmp working directory
	workDir, err := ioutil.TempDir("", "work-dir")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(workDir)

	confContent := `failover: false
replicasets:
  rpl-1:
    alias: replicaset-1
    all_rw: false
    master:
    - srv-1
    roles:
      app.roles.custom: true
      failover-coordinator: true
      vshard-router: true
    weight: 0
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: localhost:3301
  srv-expelled: expelled
`
	writeTopologyConfig(workDir, confContent)
	topologyConf, err = getTopologyConf(workDir)
	assert.Nil(err)

	replicasetConf, _ := topologyConf.Replicasets["rpl-1"]
	oldLeaders := make([]string, len(replicasetConf.Leaders))
	copy(oldLeaders, replicasetConf.Leaders)

	newInstances := []string{"srv-new-1", "srv-new-2"}
	replicasetConf.SetInstances(newInstances)
	assert.Equal(newInstances, replicasetConf.Instances)
	assert.Equal(oldLeaders, replicasetConf.Leaders)

	newContent, err = topologyConf.MarshalContent()
	assert.Nil(err)

	assert.Equal(confContent, string(newContent))
}

func TestSetLeaders(t *testing.T) {
	assert := assert.New(t)

	var err error
	var topologyConf *TopologyConfType
	var newContent []byte

	// create tmp working directory
	workDir, err := ioutil.TempDir("", "work-dir")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(workDir)

	confContent := `failover: false
replicasets:
  rpl-1:
    alias: replicaset-1
    all_rw: false
    master:
    - srv-1
    roles:
      app.roles.custom: true
      failover-coordinator: true
      vshard-router: true
    weight: 0
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: localhost:3301
  srv-expelled: expelled
`
	writeTopologyConfig(workDir, confContent)
	topologyConf, err = getTopologyConf(workDir)
	assert.Nil(err)

	replicasetConf, _ := topologyConf.Replicasets["rpl-1"]
	oldInstances := make([]string, len(replicasetConf.Instances))
	copy(oldInstances, replicasetConf.Instances)

	newLeaders := []string{"srv-new-1", "srv-new-2"}
	replicasetConf.SetLeaders(newLeaders)
	assert.Equal(newLeaders, replicasetConf.Leaders)
	assert.Equal(oldInstances, replicasetConf.Instances)

	newContent, err = topologyConf.MarshalContent()
	assert.Nil(err)

	expContent := `failover: false
replicasets:
  rpl-1:
    alias: replicaset-1
    all_rw: false
    master:
    - srv-new-1
    - srv-new-2
    roles:
      app.roles.custom: true
      failover-coordinator: true
      vshard-router: true
    weight: 0
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: localhost:3301
  srv-expelled: expelled
`

	assert.Equal(expContent, string(newContent))
}

func TestSetLeadersWhenLEadersIsString(t *testing.T) {
	assert := assert.New(t)

	var err error
	var topologyConf *TopologyConfType
	var newContent []byte

	// create tmp working directory
	workDir, err := ioutil.TempDir("", "work-dir")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(workDir)

	confContent := `failover: false
replicasets:
  rpl-1:
    alias: replicaset-1
    all_rw: false
    master: srv-1
    roles:
      app.roles.custom: true
      failover-coordinator: true
      vshard-router: true
    weight: 0
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: localhost:3301
  srv-expelled: expelled
`
	writeTopologyConfig(workDir, confContent)
	topologyConf, err = getTopologyConf(workDir)
	assert.Nil(err)

	replicasetConf, _ := topologyConf.Replicasets["rpl-1"]
	assert.True(replicasetConf.LeadersIsString)
	oldInstances := make([]string, len(replicasetConf.Instances))
	copy(oldInstances, replicasetConf.Instances)

	newLeaders := []string{"srv-new-1", "srv-new-2"}
	replicasetConf.SetLeaders(newLeaders)
	assert.Equal(newLeaders, replicasetConf.Leaders)
	assert.Equal(oldInstances, replicasetConf.Instances)

	newContent, err = topologyConf.MarshalContent()
	assert.Nil(err)

	expContent := `failover: false
replicasets:
  rpl-1:
    alias: replicaset-1
    all_rw: false
    master: srv-new-1
    roles:
      app.roles.custom: true
      failover-coordinator: true
      vshard-router: true
    weight: 0
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: localhost:3301
  srv-expelled: expelled
`

	assert.Equal(expContent, string(newContent))
}
