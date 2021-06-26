package running

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

func writeConf(file *os.File, content string) {
	if err := ioutil.WriteFile(file.Name(), []byte(content), 0644); err != nil {
		panic(fmt.Errorf("Failed to write config: %s", err))
	}
}

func TestCollectInstancesFromConfFile(t *testing.T) {
	t.Parallel()

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
	ctx.Running.ConfPath = f.Name()

	writeConf(f, `---
myapp: {}
myapp.router: {}
myapp.storage: {}
myapp-stateboard: {}
yourapp.instance: {}
`)

	instances, err = CollectInstancesFromConf(ctx)
	assert.Nil(err)
	assert.ElementsMatch(
		[]string{"router", "storage"},
		instances,
	)

	// invalid config
	ctx.Project.Name = "myapp"
	ctx.Running.ConfPath = f.Name()

	writeConf(f, `INVALID YAML`)

	instances, err = CollectInstancesFromConf(ctx)
	assert.NotNil(err)

	// non-existing file
	ctx.Project.Name = "myapp"
	ctx.Running.ConfPath = "non-existent-path"

	instances, err = CollectInstancesFromConf(ctx)
	assert.NotNil(err)
}

func TestCollectInstancesFromConfDir(t *testing.T) {
	t.Parallel()

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
	ctx.Running.ConfPath = confDirPath

	writeConf(ymlConfFile, `myapp.router: {}`)
	writeConf(yamlConfFile, `myapp.storage: {}`)
	writeConf(nonConfFile, `myapp.other: {}`)

	instances, err = CollectInstancesFromConf(ctx)
	assert.Nil(err)
	assert.ElementsMatch(
		[]string{"router", "storage"},
		instances,
	)

	// duplicate sections
	ctx.Project.Name = "myapp"
	ctx.Running.ConfPath = confDirPath

	writeConf(ymlConfFile, `myapp.router: {}`)
	writeConf(yamlConfFile, `myapp.router: {}`)
	writeConf(nonConfFile, `myapp.other: {}`)

	instances, err = CollectInstancesFromConf(ctx)
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
	t.Parallel()

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

	// Instances array also contains stateboard instance
	ctx.Project.Name = "myapp"
	ctx.Project.StateboardName = "myapp-stateboard"
	ctx.Running.WithStateboard = true
	ctx.Running.Instances = []string{"storage", "router", "stateboard"}

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
