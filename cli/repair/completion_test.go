package repair

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func writeInstanceConfig(dataDir, appName, instanceName, content string) {
	workDir := filepath.Join(dataDir, fmt.Sprintf("%s.%s", appName, instanceName))
	writeTopologyConfig(workDir, content)
}

func TestGetAllInstanceUUIDsComp(t *testing.T) {
	assert := assert.New(t)

	var err error

	// create tmp data directory
	dataDir, err := ioutil.TempDir("", "data-dir")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dataDir)

	ctx := &context.Ctx{}
	ctx.Project.Name = "myapp"
	ctx.Running.DataDir = dataDir

	instanceName1 := "srv-1"
	instanceName2 := "srv-2"

	confContent1 := `---
failover: false
replicasets:
  rpl-1:
    alias: replicaset-1
    master:
    - srv-1
    roles:
      app.roles.custom: true
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: localhost:3301
  srv-expelled: expelled
`

	confContent2 := `---
failover: false
replicasets:
  rpl-1:
    alias: replicaset-1
    master:
    - srv-1
    roles:
      app.roles.custom: true
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: localhost:3301
  srv-2:
    disabled: false
    replicaset_uuid: rpl-2
    uri: localhost:3302
`

	writeInstanceConfig(dataDir, ctx.Project.Name, instanceName1, confContent1)
	writeInstanceConfig(dataDir, ctx.Project.Name, instanceName2, confContent2)

	instanceUUIDs, err := GetAllInstanceUUIDsComp(ctx)
	assert.Nil(err)
	assert.Equal([]string{"srv-1", "srv-2", "srv-expelled"}, instanceUUIDs)
}

func TestGetInstanceHostsComp(t *testing.T) {
	assert := assert.New(t)

	var err error

	// create tmp data directory
	dataDir, err := ioutil.TempDir("", "data-dir")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dataDir)

	ctx := &context.Ctx{}
	ctx.Project.Name = "myapp"
	ctx.Running.DataDir = dataDir

	instanceName1 := "srv-1"
	instanceName2 := "srv-2"

	confContent1 := `---
failover: false
replicasets:
  rpl-1:
    alias: replicaset-1
    master:
    - srv-1
    roles:
      app.roles.custom: true
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: localhost:3301
  srv-expelled: expelled
`

	confContent2 := `---
failover: false
replicasets:
  rpl-1:
    alias: replicaset-1
    master:
    - srv-1
    roles:
      app.roles.custom: true
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: globalhost:3301
`

	writeInstanceConfig(dataDir, ctx.Project.Name, instanceName1, confContent1)
	writeInstanceConfig(dataDir, ctx.Project.Name, instanceName2, confContent2)

	instanceHosts, err := GetInstanceHostsComp("srv-1", ctx)
	assert.Nil(err)
	assert.Equal([]string{"globalhost", "localhost"}, instanceHosts)
}

func TestGetAllReplicasetUUIDsComp(t *testing.T) {
	assert := assert.New(t)

	var err error

	// create tmp data directory
	dataDir, err := ioutil.TempDir("", "data-dir")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dataDir)

	ctx := &context.Ctx{}
	ctx.Project.Name = "myapp"
	ctx.Running.DataDir = dataDir

	instanceName1 := "srv-1"
	instanceName2 := "srv-2"

	confContent1 := `---
failover: false
replicasets:
  rpl-1:
    alias: replicaset-1
    master:
    - srv-1
    roles:
      app.roles.custom: true
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: localhost:3301
  srv-expelled: expelled
`

	confContent2 := `---
failover: false
replicasets:
  rpl-2:
    alias: replicaset-2
    master:
    - srv-2
    roles:
      app.roles.custom: true
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: localhost:3301
  srv-2:
    disabled: false
    replicaset_uuid: rpl-2
    uri: localhost:3302
`

	writeInstanceConfig(dataDir, ctx.Project.Name, instanceName1, confContent1)
	writeInstanceConfig(dataDir, ctx.Project.Name, instanceName2, confContent2)

	replicasetUUIDs, err := GetAllReplicasetUUIDsComp(ctx)
	assert.Nil(err)
	assert.Equal([]string{"rpl-1", "rpl-2"}, replicasetUUIDs)
}

func TestGetReplicasetInstancesComp(t *testing.T) {
	assert := assert.New(t)

	var err error

	// create tmp data directory
	dataDir, err := ioutil.TempDir("", "data-dir")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dataDir)

	ctx := &context.Ctx{}
	ctx.Project.Name = "myapp"
	ctx.Running.DataDir = dataDir

	instanceName1 := "srv-1"
	instanceName2 := "srv-2"

	confContent1 := `---
failover: false
replicasets:
  rpl-1:
    alias: replicaset-1
    master:
    - srv-1
    roles:
      app.roles.custom: true
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: localhost:3301
  srv-2:
    disabled: false
    replicaset_uuid: rpl-2
    uri: localhost:3302
  srv-expelled: expelled
`

	confContent2 := `---
failover: false
replicasets:
  rpl-1:
    alias: replicaset-1
    master:
    - srv-1
    roles:
      app.roles.custom: true
servers:
  srv-1:
    disabled: false
    replicaset_uuid: rpl-1
    uri: localhost:3301
  srv-2:
    disabled: false
    replicaset_uuid: rpl-2
    uri: localhost:3302
  srv-3:
    disabled: true
    replicaset_uuid: rpl-1
    uri: localhost:3302
  srv-expelled: expelled
`

	writeInstanceConfig(dataDir, ctx.Project.Name, instanceName1, confContent1)
	writeInstanceConfig(dataDir, ctx.Project.Name, instanceName2, confContent2)

	replicasetUUIDs, err := GetReplicasetInstancesComp("rpl-1", ctx)
	assert.Nil(err)
	assert.Equal([]string{"srv-1", "srv-3"}, replicasetUUIDs)
}
