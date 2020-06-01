package pack

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tarantool/cartridge-cli/cli/project"
)

func TestGetPackageFullname(t *testing.T) {
	assert := assert.New(t)

	var projectCtx project.ProjectCtx

	// TODO: internal error on bad type

	// w/o suffix
	projectCtx.Name = "myapp"
	projectCtx.VersionRelease = "1.2.3-4"
	projectCtx.Suffix = ""

	projectCtx.PackType = tgzType
	assert.Equal("myapp-1.2.3-4.tar.gz", getPackageFullname(&projectCtx))

	projectCtx.PackType = rpmType
	assert.Equal("myapp-1.2.3-4.rpm", getPackageFullname(&projectCtx))

	projectCtx.PackType = debType
	assert.Equal("myapp-1.2.3-4.deb", getPackageFullname(&projectCtx))

	// w/ suffix
	projectCtx.Name = "myapp"
	projectCtx.VersionRelease = "1.2.3-4"
	projectCtx.Suffix = "dev"

	projectCtx.PackType = tgzType
	assert.Equal("myapp-1.2.3-4-dev.tar.gz", getPackageFullname(&projectCtx))

	projectCtx.PackType = rpmType
	assert.Equal("myapp-1.2.3-4-dev.rpm", getPackageFullname(&projectCtx))

	projectCtx.PackType = debType
	assert.Equal("myapp-1.2.3-4-dev.deb", getPackageFullname(&projectCtx))
}

func TestGetImageFullname(t *testing.T) {
	assert := assert.New(t)

	var projectCtx project.ProjectCtx

	// TODO: internal error on bad type

	// VersionRelease
	projectCtx.Name = "myapp"
	projectCtx.VersionRelease = "1.2.3-4"
	projectCtx.Suffix = ""
	projectCtx.ImageTag = ""

	assert.Equal("myapp:1.2.3-4", getImageFullname(&projectCtx))

	// VersionRelease + Suffix
	projectCtx.Name = "myapp"
	projectCtx.VersionRelease = "1.2.3-4"
	projectCtx.Suffix = "dev"
	projectCtx.ImageTag = ""

	assert.Equal("myapp:1.2.3-4-dev", getImageFullname(&projectCtx))

	// ImageTag
	projectCtx.Name = "myapp"
	projectCtx.VersionRelease = ""
	projectCtx.Suffix = ""
	projectCtx.ImageTag = "my-first-image"

	assert.Equal("my-first-image", getImageFullname(&projectCtx))
}
