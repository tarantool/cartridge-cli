package templates

import (
	"github.com/tarantool/cartridge-cli/cli/templates"
)

var buildFilesTemplate = templates.FileTreeTemplate{
	Dirs: []templates.DirTemplate{},
	Files: []templates.FileTemplate{
		{
			Path:    "cartridge.pre-build",
			Mode:    0755,
			Content: GetStaticFileContent("build/cartridge.pre-build"),
		},
		{
			Path:    "cartridge.post-build",
			Mode:    0755,
			Content: GetStaticFileContent("build/cartridge.post-build"),
		},
		{
			Path:    "Dockerfile.build.cartridge",
			Mode:    0644,
			Content: GetStaticFileContent("build/Dockerfile.build.cartridge"),
		},
		{
			Path:    "Dockerfile.cartridge",
			Mode:    0644,
			Content: GetStaticFileContent("build/Dockerfile.cartridge"),
		},
	},
}
