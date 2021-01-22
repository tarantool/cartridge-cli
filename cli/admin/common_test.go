package admin

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

func TestCheckCtx(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	var ctx context.Ctx

	ctx.Project.Name = ""
	ctx.Admin.ConnString = ""
	ctx.Admin.InstanceName = ""
	assert.EqualError(checkCtx(&ctx), "Please, specify one of --name, --instance or --conn")

	ctx.Project.Name = "app-name"
	ctx.Admin.ConnString = "conn-string"
	ctx.Admin.InstanceName = "instance-name"
	assert.EqualError(checkCtx(&ctx), "You can specify only one of --instance or --conn")

	ctx.Project.Name = ""
	ctx.Admin.ConnString = "conn-string"
	ctx.Admin.InstanceName = "instance-name"
	assert.EqualError(checkCtx(&ctx), "You can specify only one of --instance or --conn")

	ctx.Project.Name = ""
	ctx.Admin.ConnString = ""
	ctx.Admin.InstanceName = "instance-name"
	assert.EqualError(checkCtx(&ctx), "Please, specify --name")

	ctx.Project.Name = "app-name"
	ctx.Admin.ConnString = ""
	ctx.Admin.InstanceName = ""
	assert.Nil(checkCtx(&ctx))

	ctx.Project.Name = "app-name"
	ctx.Admin.ConnString = ""
	ctx.Admin.InstanceName = "instance-name"
	assert.Nil(checkCtx(&ctx))

	ctx.Project.Name = ""
	ctx.Admin.ConnString = "conn-string"
	ctx.Admin.InstanceName = ""
	assert.Nil(checkCtx(&ctx))
}

func cleanDir(dirPath string) {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if err := os.Remove(filepath.Join(dirPath, file.Name())); err != nil {
			log.Fatal(err)
		}
	}
}

func createFiles(dirPath string, fileNames []string) {
	for _, fileName := range fileNames {
		filePath := filepath.Join(dirPath, fileName)
		if _, err := os.Create(filePath); err != nil {
			log.Fatal(err)
		}
	}
}

func TestGetInstanceSocketPaths(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	var ctx context.Ctx
	var err error
	var addresses []string

	runDir, err := ioutil.TempDir("", "run")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(runDir)

	// run-dir is empty
	cleanDir(runDir)
	ctx.Project.Name = "myapp"
	ctx.Running.RunDir = runDir
	addresses, err = getInstanceSocketPaths(&ctx)
	assert.Nil(addresses)
	assert.EqualError(err, fmt.Sprintf("Run directory %s is empty", runDir))

	// no control sockets
	cleanDir(runDir)
	createFiles(runDir, []string{
		"myapp.router.pid",
		"myapp.router.notify",
		"some-wtf-file",
	})

	ctx.Project.Name = "myapp"
	ctx.Running.RunDir = runDir
	addresses, err = getInstanceSocketPaths(&ctx)
	assert.Nil(addresses)
	assert.EqualError(err, fmt.Sprintf("No instance sockets found in %s", runDir))

	// control sockets of different apps and other files
	cleanDir(runDir)
	createFiles(runDir, []string{
		"myapp.router.pid",
		"myapp.router.notify",
		"myapp.router.control",
		"myapp.storage.pid",
		"myapp.storage.notify",
		"myapp.storage.control",
		"other-app.router.pid",
		"other-app.router.notify",
		"other-app.router.control",
		"some-wtf-file",
	})

	ctx.Project.Name = "myapp"
	ctx.Running.RunDir = runDir
	addresses, err = getInstanceSocketPaths(&ctx)
	assert.Nil(err)
	assert.EqualValues(addresses, []string{
		filepath.Join(runDir, "myapp.router.control"),
		filepath.Join(runDir, "myapp.storage.control"),
	})
}
