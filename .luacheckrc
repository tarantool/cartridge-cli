redefined = false
exclude_files = {}
include_files = {'**/*.lua', '*.luacheckrc', 'cartridge'}
new_read_globals = {
    '_TARANTOOL',
    'box',
    package = {
      fields = {
        'search',
        'setsearchroot',
      }
    },
    string = {
        fields = {
            'split',
            'strip',
            'startswith',
            'endswith',
        },
    },
}
