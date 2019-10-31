package = 'cartridge-cli'
version = 'scm-1'
source = {
    url = 'git+https://github.com/rosik/cartridge-cli.git',
    branch = 'master',
}

dependencies = {
    'lua >= 5.1',
}

build = {
    type = 'cmake',

    variables = {
        TARANTOOL_DIR = '$(TARANTOOL_DIR)',
        TARANTOOL_INSTALL_LIBDIR = '$(LIBDIR)',
        TARANTOOL_INSTALL_LUADIR = '$(LUADIR)',
        TARANTOOL_INSTALL_BINDIR = '$(BINDIR)',
    }
}
