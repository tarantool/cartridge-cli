local config_files = {
    {
        name = '.luacheckrc',
        mode = tonumber('0644', 8),
        content = [=[
            include_files = {'**/*.lua', '*.luacheckrc', '*.rockspec'}
            exclude_files = {'.rocks/', 'tmp/'}
            max_line_length = 120
        ]=]
    },
    {
        name = '.luacov',
        mode = tonumber('0644', 8),
        content = [=[
            statsfile = 'tmp/luacov.stats.out'
            reportfile = 'tmp/luacov.report.out'
            exclude = {
                '/test/',
            }
        ]=]
    },
    {
        name = '.editorconfig',
        mode = tonumber('0644', 8),
        content = [=[
            # top-most EditorConfig file
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
        ]=]
    }
}

return config_files
