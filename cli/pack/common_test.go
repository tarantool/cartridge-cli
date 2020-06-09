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

	projectCtx.PackType = TgzType
	assert.Equal("myapp-1.2.3-4.tar.gz", getPackageFullname(&projectCtx))

	projectCtx.PackType = RpmType
	assert.Equal("myapp-1.2.3-4.rpm", getPackageFullname(&projectCtx))

	projectCtx.PackType = DebType
	assert.Equal("myapp-1.2.3-4.deb", getPackageFullname(&projectCtx))

	// w/ suffix
	projectCtx.Name = "myapp"
	projectCtx.VersionRelease = "1.2.3-4"
	projectCtx.Suffix = "dev"

	projectCtx.PackType = TgzType
	assert.Equal("myapp-1.2.3-4-dev.tar.gz", getPackageFullname(&projectCtx))

	projectCtx.PackType = RpmType
	assert.Equal("myapp-1.2.3-4-dev.rpm", getPackageFullname(&projectCtx))

	projectCtx.PackType = DebType
	assert.Equal("myapp-1.2.3-4-dev.deb", getPackageFullname(&projectCtx))
}

func TestGetImageTags(t *testing.T) {
	assert := assert.New(t)

	var projectCtx project.ProjectCtx

	// TODO: internal error on bad type

	// VersionRelease
	projectCtx.Name = "myapp"
	projectCtx.VersionRelease = "1.2.3-4"
	projectCtx.Suffix = ""
	projectCtx.ImageTags = []string{}

	assert.ElementsMatch([]string{"myapp:1.2.3-4"}, getImageTags(&projectCtx))

	// VersionRelease + Suffix
	projectCtx.Name = "myapp"
	projectCtx.VersionRelease = "1.2.3-4"
	projectCtx.Suffix = "dev"
	projectCtx.ImageTags = []string{}

	assert.ElementsMatch([]string{"myapp:1.2.3-4-dev"}, getImageTags(&projectCtx))

	// ImageTags
	projectCtx.Name = "myapp"
	projectCtx.VersionRelease = ""
	projectCtx.Suffix = ""
	projectCtx.ImageTags = []string{"my-first-image", "my-lovely-image"}

	assert.ElementsMatch([]string{"my-first-image", "my-lovely-image"}, getImageTags(&projectCtx))
}
