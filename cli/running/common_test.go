package running

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tarantool/cartridge-cli/cli/project"

	"github.com/stretchr/testify/assert"
)

func writeConf(file *os.File, content string) {
	if err := ioutil.WriteFile(file.Name(), []byte(content), 0644); err != nil {
		panic(fmt.Errorf("Failed to write config: %s", err))
	}
}

func TestCollectInstancesFromConfFile(t *testing.T) {
	assert := assert.New(t)

	var err error
	var instances []string

	ctx := &project.ProjectCtx{}

	// create tmp conf file
	f, err := ioutil.TempFile("", "myapp.yml")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	// valid config
	ctx.Name = "myapp"
	ctx.ConfPath = f.Name()

	writeConf(f, `---
myapp: {}
myapp.router: {}
myapp.storage: {}
myapp-stateboard: {}
yourapp.instance: {}
`)

	instances, err = collectInstancesFromConf(ctx)
	assert.Nil(err)
	assert.ElementsMatch(
		[]string{"router", "storage"},
		instances,
	)

	// invalid config
	ctx.Name = "myapp"
	ctx.ConfPath = f.Name()

	writeConf(f, `INVALID YAML`)

	instances, err = collectInstancesFromConf(ctx)
	assert.NotNil(err)

	// non-existing file
	ctx.Name = "myapp"
	ctx.ConfPath = "non-existent-path"

	instances, err = collectInstancesFromConf(ctx)
	assert.NotNil(err)
}

func TestCollectInstancesFromConfDir(t *testing.T) {
	assert := assert.New(t)

	var err error
	var instances []string

	ctx := &project.ProjectCtx{}

	// create tmp conf dir
	confDirPath, err := ioutil.TempDir("", "myapp_conf")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(confDirPath)

	// create two config files and one none:
	// myapp.yml, myapp.yaml, some-other-file
	ymlConfFile, err := os.Create(filepath.Join(confDirPath, "myapp.yml"))
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(ymlConfFile.Name())

	yamlConfFile, err := os.Create(filepath.Join(confDirPath, "myapp.yaml"))
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(yamlConfFile.Name())

	nonConfFile, err := os.Create(filepath.Join(confDirPath, "some-file"))
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(nonConfFile.Name())

	// valid config
	ctx.Name = "myapp"
	ctx.ConfPath = confDirPath

	writeConf(ymlConfFile, `myapp.router: {}`)
	writeConf(yamlConfFile, `myapp.storage: {}`)
	writeConf(nonConfFile, `myapp.other: {}`)

	instances, err = collectInstancesFromConf(ctx)
	assert.Nil(err)
	assert.ElementsMatch(
		[]string{"router", "storage"},
		instances,
	)

	// duplicate sections
	ctx.Name = "myapp"
	ctx.ConfPath = confDirPath

	writeConf(ymlConfFile, `myapp.router: {}`)
	writeConf(yamlConfFile, `myapp.router: {}`)
	writeConf(nonConfFile, `myapp.other: {}`)

	instances, err = collectInstancesFromConf(ctx)
	assert.NotNil(err)
}

func getProcessesIDs(processes *ProcessesSet) []string {
	var ids []string

	for _, process := range *processes {
		ids = append(ids, process.ID)
	}

	return ids
}

func TestCollectProcesses(t *testing.T) {
	assert := assert.New(t)

	var err error
	var processes *ProcessesSet

	ctx := &project.ProjectCtx{}

	// project w/ stateboard
	ctx.Name = "myapp"
	ctx.StateboardName = "myapp-stateboard"
	ctx.WithStateboard = true
	ctx.Instances = []string{"storage", "router"}

	processes, err = collectProcesses(ctx)
	assert.Nil(err)
	assert.ElementsMatch(
		[]string{"myapp.router", "myapp.storage", "myapp-stateboard"},
		getProcessesIDs(processes),
	)

	// project w/o stateboard
	ctx.Name = "myapp"
	ctx.StateboardName = "myapp-stateboard"
	ctx.WithStateboard = false
	ctx.Instances = []string{"storage", "router"}

	processes, err = collectProcesses(ctx)
	assert.Nil(err)
	assert.ElementsMatch(
		[]string{"myapp.router", "myapp.storage"},
		getProcessesIDs(processes),
	)

	// stateboard only
	ctx.Name = "myapp"
	ctx.StateboardName = "myapp-stateboard"
	ctx.WithStateboard = true
	ctx.StateboardOnly = true
	ctx.Instances = []string{"storage", "router"}

	processes, err = collectProcesses(ctx)
	assert.Nil(err)
	assert.ElementsMatch(
		[]string{"myapp-stateboard"},
		getProcessesIDs(processes),
	)
}

func TestGetInstancesFromArgs(t *testing.T) {
	assert := assert.New(t)

	var err error
	var instanceIDs []string
	var instances []string

	ctx := &project.ProjectCtx{}
	ctx.Name = "myapp"

	// wrong format
	instanceIDs = []string{"myapp.instance.1"}
	_, err = GetInstancesFromArgs(instanceIDs, ctx)
	assert.EqualError(err, "Instance ID should be [APP_NAME][.INSTANCE_NAME]")

	// wrong application name
	instanceIDs = []string{"wrong-name.instance-1"}
	_, err = GetInstancesFromArgs(instanceIDs, ctx)
	assert.True(strings.Contains(err.Error(), "Wrong application name: wrong-name"))

	instanceIDs = []string{"wrong-name"}
	_, err = GetInstancesFromArgs(instanceIDs, ctx)
	assert.True(strings.Contains(err.Error(), "Wrong application name: wrong-name"))

	// duplicate instance name
	instanceIDs = []string{"myapp.instance-1", ".instance-1"}
	_, err = GetInstancesFromArgs(instanceIDs, ctx)
	assert.True(strings.Contains(err.Error(), "Duplicate instance name: instance-1"))

	// instances are specified
	instanceIDs = []string{".instance-1", ".instance-2"}
	instances, err = GetInstancesFromArgs(instanceIDs, ctx)
	assert.Nil(err)
	assert.Equal([]string{"instance-1", "instance-2"}, instances)

	instanceIDs = []string{"myapp.instance-1", "myapp.instance-2"}
	instances, err = GetInstancesFromArgs(instanceIDs, ctx)
	assert.Nil(err)
	assert.Equal([]string{"instance-1", "instance-2"}, instances)

	instanceIDs = []string{".instance-1", "myapp.instance-2"}
	instances, err = GetInstancesFromArgs(instanceIDs, ctx)
	assert.Nil(err)
	assert.Equal([]string{"instance-1", "instance-2"}, instances)

	// appname specified more than one time
	instanceIDs = []string{"myapp", "myapp"}
	instances, err = GetInstancesFromArgs(instanceIDs, ctx)
	assert.EqualError(err, specifyAppOrInstancesErr)

	// specified both app name and instance ID
	instanceIDs = []string{"myapp.instance-1", "myapp"}
	instances, err = GetInstancesFromArgs(instanceIDs, ctx)
	assert.EqualError(err, specifyAppOrInstancesErr)

	instanceIDs = []string{"myapp", "myapp.instance-1"}
	instances, err = GetInstancesFromArgs(instanceIDs, ctx)
	assert.EqualError(err, specifyAppOrInstancesErr)
}
