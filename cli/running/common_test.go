package running

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarantool/cartridge-cli/cli/context"
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

	ctx := &context.Ctx{}

	// create tmp conf file
	f, err := ioutil.TempFile("", "myapp.yml")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	// valid config
	ctx.Project.Name = "myapp"
	ctx.Running.AppConfPath = f.Name()

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
	ctx.Project.Name = "myapp"
	ctx.Running.AppConfPath = f.Name()

	writeConf(f, `INVALID YAML`)

	instances, err = collectInstancesFromConf(ctx)
	assert.NotNil(err)

	// non-existing file
	ctx.Project.Name = "myapp"
	ctx.Running.AppConfPath = "non-existent-path"

	instances, err = collectInstancesFromConf(ctx)
	assert.NotNil(err)
}

func TestCollectInstancesFromConfDir(t *testing.T) {
	assert := assert.New(t)

	var err error
	var instances []string

	ctx := &context.Ctx{}

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
	ctx.Project.Name = "myapp"
	ctx.Running.AppConfPath = confDirPath

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
	ctx.Project.Name = "myapp"
	ctx.Running.AppConfPath = confDirPath

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

	ctx := &context.Ctx{}

	// project w/ stateboard
	ctx.Project.Name = "myapp"
	ctx.Project.StateboardName = "myapp-stateboard"
	ctx.Running.WithStateboard = true
	ctx.Running.Instances = []string{"storage", "router"}

	processes, err = collectProcesses(ctx)
	assert.Nil(err)
	assert.ElementsMatch(
		[]string{"myapp.router", "myapp.storage", "myapp-stateboard"},
		getProcessesIDs(processes),
	)

	// project w/o stateboard
	ctx.Project.Name = "myapp"
	ctx.Project.StateboardName = "myapp-stateboard"
	ctx.Running.WithStateboard = false
	ctx.Running.Instances = []string{"storage", "router"}

	processes, err = collectProcesses(ctx)
	assert.Nil(err)
	assert.ElementsMatch(
		[]string{"myapp.router", "myapp.storage"},
		getProcessesIDs(processes),
	)

	// stateboard only
	ctx.Project.Name = "myapp"
	ctx.Project.StateboardName = "myapp-stateboard"
	ctx.Running.WithStateboard = true
	ctx.Running.StateboardOnly = true
	ctx.Running.Instances = []string{"storage", "router"}

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
	var args []string
	var instances []string

	ctx := &context.Ctx{}
	ctx.Project.Name = "myapp"

	// wrong format
	args = []string{"myapp.instance-1", "myapp.instance-2"}
	_, err = getInstancesFromArgs(args, ctx)
	assert.EqualError(err, instanceIDSpecified)

	args = []string{"instance-1", "myapp.instance-2"}
	_, err = getInstancesFromArgs(args, ctx)
	assert.EqualError(err, instanceIDSpecified)

	args = []string{"myapp"}
	_, err = getInstancesFromArgs(args, ctx)
	assert.True(strings.Contains(err.Error(), appNameSpecifiedError))

	// duplicate instance name
	args = []string{"instance-1", "instance-1"}
	_, err = getInstancesFromArgs(args, ctx)
	assert.True(strings.Contains(err.Error(), "Duplicate instance name: instance-1"))

	// instances are specified
	args = []string{"instance-1", "instance-2"}
	instances, err = getInstancesFromArgs(args, ctx)
	assert.Nil(err)
	assert.Equal([]string{"instance-1", "instance-2"}, instances)

	// specified both app name and instance name
	args = []string{"instance-1", "myapp"}
	instances, err = getInstancesFromArgs(args, ctx)
	assert.EqualError(err, appNameSpecifiedError)

	args = []string{"myapp", "instance-1"}
	instances, err = getInstancesFromArgs(args, ctx)
	assert.EqualError(err, appNameSpecifiedError)
}
