package templates

var configFilesTemplate = projectTemplate{
	Dirs:  getCartridgeConfigDirs(),
	Files: getCartridgeConfigFiles(),
}

func getCartridgeConfigDirs() []dirTemplate {
	return []dirTemplate{}
}

func getCartridgeConfigFiles() []fileTemplate {
	luacheckrc := fileTemplate{
		Path: ".luacheckrc",
		Mode: 0644,
		Content: `include_files = {'**/*.lua', '*.luacheckrc', '*.rockspec'}
exclude_files = {'.rocks/', 'tmp/'}
max_line_length = 120
`,
	}

	luacov := fileTemplate{
		Path: ".luacov",
		Mode: 0644,
		Content: `
statsfile = 'tmp/luacov.stats.out'
reportfile = 'tmp/luacov.report.out'
exclude = {
	'/test/',
}
`,
	}

	editorconfig := fileTemplate{
		Path: ".editorconfig",
		Mode: 0644,
		Content: `# top-most EditorConfig file
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
`,
	}

	return []fileTemplate{
		luacheckrc,
		luacov,
		editorconfig,
	}
}
