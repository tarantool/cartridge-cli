package templates

import "github.com/tarantool/cartridge-cli/cli/templates"

var testFilesTemplate = templates.FileTreeTemplate{
	Dirs: []templates.DirTemplate{
		{
			Path: "test/helper",
			Mode: 0755,
		},
		{
			Path: "test/integration",
			Mode: 0755,
		},
		{
			Path: "test/unit",
			Mode: 0755,
		},
	},
	Files: []templates.FileTemplate{
		{
			Path:    "test/helper/integration.lua",
			Mode:    0644,
			Content: GetStaticFileContent("test/integration.lua"),
		},

		{
			Path:    "test/helper/unit.lua",
			Mode:    0644,
			Content: GetStaticFileContent("test/unit.lua"),
		},

		{
			Path:    "test/helper.lua",
			Mode:    0644,
			Content: GetStaticFileContent("test/helper.lua"),
		},

		{
			Path:    "test/integration/api_test.lua",
			Mode:    0644,
			Content: GetStaticFileContent("test/api_test.lua"),
		},

		{
			Path:    "test/unit/sample_test.lua",
			Mode:    0644,
			Content: GetStaticFileContent("test/sample_test.lua"),
		},
	},
}
