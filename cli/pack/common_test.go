package pack

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tarantool/cartridge-cli/cli/context"
)

var nameTests = map[string]struct {
	// Input
	name     string
	version  string
	suffix   string
	packType string
	// Output
	packageFullname string
}{
	"X.Y.Z_no_suffix_rpm": {
		name:            "myapp",
		version:         "1.2.3",
		suffix:          "",
		packType:        RpmType,
		packageFullname: "myapp-1.2.3.0-1.x86_64.rpm",
	},
	"X.Y.Z_no_suffix_deb": {
		name:            "myapp",
		version:         "1.2.3",
		suffix:          "",
		packType:        DebType,
		packageFullname: "myapp_1.2.3.0-1_all.deb",
	},
	"X.Y.Z_no_suffix_tgz": {
		name:            "myapp",
		version:         "1.2.3",
		suffix:          "",
		packType:        TgzType,
		packageFullname: "myapp-1.2.3.0.x86_64.tar.gz",
	},
	"X.Y.Z-N_no_suffix_rpm": {
		name:            "myapp",
		version:         "1.2.3-4",
		suffix:          "",
		packType:        RpmType,
		packageFullname: "myapp-1.2.3.4-1.x86_64.rpm",
	},
	"X.Y.Z-N_no_suffix_deb": {
		name:            "myapp",
		version:         "1.2.3-4",
		suffix:          "",
		packType:        DebType,
		packageFullname: "myapp_1.2.3.4-1_all.deb",
	},
	"X.Y.Z-N_no_suffix_tgz": {
		name:            "myapp",
		version:         "1.2.3-4",
		suffix:          "",
		packType:        TgzType,
		packageFullname: "myapp-1.2.3.4.x86_64.tar.gz",
	},
	"X.Y.Z-N-ghash_no_suffix_rpm": {
		name:            "myapp",
		version:         "1.2.3-4-g480c55b67",
		suffix:          "",
		packType:        RpmType,
		packageFullname: "myapp-1.2.3.4-1.x86_64.rpm",
	},
	"X.Y.Z-N-ghash_no_suffix_deb": {
		name:            "myapp",
		version:         "1.2.3-4-g480c55b67",
		suffix:          "",
		packType:        DebType,
		packageFullname: "myapp_1.2.3.4-1_all.deb",
	},
	"X.Y.Z-N-ghash_no_suffix_tgz": {
		name:            "myapp",
		version:         "1.2.3-4-g480c55b67",
		suffix:          "",
		packType:        TgzType,
		packageFullname: "myapp-1.2.3.4.x86_64.tar.gz",
	},
	"X.Y.Z-N-ghash_with_suffix_rpm": {
		name:            "myapp",
		version:         "1.2.3-4-g480c55b67",
		suffix:          "dev",
		packType:        RpmType,
		packageFullname: "myapp-1.2.3.4.dev-1.x86_64.rpm",
	},
	"X.Y.Z-N-ghash_with_suffix_deb": {
		name:            "myapp",
		version:         "1.2.3-4-g480c55b67",
		suffix:          "dev",
		packType:        DebType,
		packageFullname: "myapp_1.2.3.4.dev-1_all.deb",
	},
	"X.Y.Z-N-ghash_with_suffix_tgz": {
		name:            "myapp",
		version:         "1.2.3-4-g480c55b67",
		suffix:          "dev",
		packType:        TgzType,
		packageFullname: "myapp-1.2.3.4.dev.x86_64.tar.gz",
	},
}

func TestGetPackageFullname(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	for tname, tt := range nameTests {
		t.Run(tname, func(t *testing.T) {
			var ctx context.Ctx

			ctx.Project.Name = tt.name
			ctx.Pack.Version = tt.version
			ctx.Pack.Suffix = tt.suffix
			ctx.Pack.Type = tt.packType

			assert.Equal(nil, normalizeGitVersion(&ctx))
			assert.Equal(nil, buildVersionWithSuffix(&ctx))
			detectRelease(&ctx)
			detectArch(&ctx)

			assert.Equal(tt.packageFullname, getPackageFullname(&ctx))
		})
	}
}

func TestGetImageTags(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	var ctx context.Ctx

	// TODO: internal error on bad type

	// VersionRelease
	ctx.Project.Name = "myapp"
	ctx.Pack.Version = "1.2.3-4"
	ctx.Pack.Suffix = ""
	ctx.Pack.ImageTags = []string{}

	assert.ElementsMatch([]string{"myapp:1.2.3-4"}, getImageTags(&ctx))

	// VersionRelease + Suffix
	ctx.Project.Name = "myapp"
	ctx.Pack.Version = "1.2.3-4"
	ctx.Pack.Suffix = "dev"
	ctx.Pack.ImageTags = []string{}

	assert.ElementsMatch([]string{"myapp:1.2.3-4-dev"}, getImageTags(&ctx))

	// ImageTags
	ctx.Project.Name = "myapp"
	ctx.Pack.Version = ""
	ctx.Pack.Suffix = ""
	ctx.Pack.ImageTags = []string{"my-first-image", "my-lovely-image"}

	assert.ElementsMatch([]string{"my-first-image", "my-lovely-image"}, getImageTags(&ctx))
}
