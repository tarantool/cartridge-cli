package templates

import "github.com/tarantool/cartridge-cli/cli/templates"

var devFilesTemplate = templates.FileTreeTemplate{
	Dirs: []templates.DirTemplate{
		{
			Path: "tmp",
			Mode: 0755,
		},
	},
	Files: []templates.FileTemplate{
		{
			Path:    "deps.sh",
			Mode:    0755,
			Content: GetStaticFileContent("dev/deps.sh"),
		},
		{
			Path:    "instances.yml",
			Mode:    0644,
			Content: GetStaticFileContent("dev/instances.yml"),
		},
		{
			Path:    "replicasets.yml",
			Mode:    0644,
			Content: GetStaticFileContent("dev/replicasets.yml"),
		},
		{
			Path:    ".cartridge.yml",
			Mode:    0644,
			Content: GetStaticFileContent("dev/.cartridge.yml"),
		},
		{
			Path:    "tmp/.keep",
			Mode:    0644,
			Content: "",
		},
	},
}
