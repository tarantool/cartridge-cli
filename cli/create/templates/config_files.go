package templates

import "github.com/tarantool/cartridge-cli/cli/templates"

var configFilesTemplate = templates.FileTreeTemplate{
	Dirs: []templates.DirTemplate{},
	Files: []templates.FileTemplate{
		{
			Path:    ".luacheckrc",
			Mode:    0644,
			Content: luacheckrcContent,
		},
		{
			Path:    ".luacov",
			Mode:    0644,
			Content: luacovContent,
		},
		{
			Path:    ".editorconfig",
			Mode:    0644,
			Content: editorconfigContent,
		},
		{
			Path:    ".gitignore",
			Mode:    0644,
			Content: gitIgnoreContent,
		},
	},
}

const (
	luacheckrcContent = `include_files = {'**/*.lua', '*.luacheckrc', '*.rockspec'}
exclude_files = {'.rocks/', 'tmp/'}
max_line_length = 120
`
	luacovContent = `
statsfile = 'tmp/luacov.stats.out'
reportfile = 'tmp/luacov.report.out'
exclude = {
	'/test/',
}
`
	editorconfigContent = `# top-most EditorConfig file
root = true
# Unix-style newlines with a newline ending every file
[*]
end_of_line = lf
insert_final_newline = true
[CMakeLists.txt]
indent_style = space
indent_size = 4
[*.cmake]
indent_style = space
indent_size = 4
[*.lua]
indent_style = space
indent_size = 4
[*.{h,c,cc}]
indent_style = tab
tab_width = 8
`

	gitIgnoreContent = `.rocks
.swo
.swp
CMakeCache.txt
CMakeFiles
cmake_install.cmake
*.dylib
*.idea
__pycache__
*pyc
.cache
.pytest_cache
.vagrant
.DS_Store
*.xlog
*.snap
*.rpm
*.deb
*.tar.gz
node_modules
/tmp/*
!/tmp/.keep
`
)
