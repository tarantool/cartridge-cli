package pack

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func TestGenerateVersionFileNameEE(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	var ctx context.Ctx

	ctx.Project.Name = "myapp"
	ctx.Pack.VersionRelease = "123"
	ctx.Tarantool.TarantoolIsEnterprise = true
	ctx.Build.InDocker = true

	dir, err := ioutil.TempDir("", "__temporary_sdk")
	if err != nil {
		assert.Equal(err, nil)
	}
	defer os.RemoveAll(dir)

	ctx.Build.SDKPath = dir
	versionFileLines := []string{
		"TARANTOOL=2.8.1-0-ge2a1ec0c2-r409",
		"TARANTOOL_SDK=2.8.1-0-ge2a1ec0c2-r409",
	}

	tmpVersion := filepath.Join(dir, "VERSION")
	if err := ioutil.WriteFile(tmpVersion, []byte(strings.Join(versionFileLines, "\n")), 0666); err != nil {
		assert.Equal(nil, err)
	}

	err = generateVersionFile("", &ctx)
	assert.Equal(nil, err)

	content, err := ioutil.ReadFile("VERSION")
	if err != nil {
		assert.Equal(nil, err)
	}

	for i, line := range strings.Split(string(content), "\n")[:3] {
		if i == 0 { // app name
			assert.Equal(fmt.Sprintf("%s=%s", ctx.Project.Name, ctx.Pack.VersionRelease), line)
		} else { // tarantool versions
			assert.Equal(versionFileLines[i-1], line)
		}
	}

	os.Remove("VERSION")
}
