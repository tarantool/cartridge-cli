package templates

import "github.com/tarantool/cartridge-cli/cli/templates"

var configFilesTemplate = templates.FileTreeTemplate{
	Dirs: []templates.DirTemplate{},
	Files: []templates.FileTemplate{
		{
			Path:    ".luacheckrc",
			Mode:    0644,
			Content: GetStaticFileContent("config/.luacheckrc"),
		},
		{
			Path:    ".luacov",
			Mode:    0644,
			Content: GetStaticFileContent("config/.luacov"),
		},
		{
			Path:    ".editorconfig",
			Mode:    0644,
			Content: GetStaticFileContent("config/.editorconfig"),
		},
		{
			Path:    ".gitignore",
			Mode:    0644,
			Content: GetStaticFileContent("config/.gitignore"),
		},
	},
}
