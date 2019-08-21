package = 'errors'
version = '2.1.1-1'
source = {
    url = 'git+https://github.com/tarantool/errors.git',
    branch = 'master',
}

description = {
    summary = 'Convenient error handling in tarantool',
    homepage = 'https://github.com/tarantool/errors',
    license = 'BSD',
}

dependencies = {
    'lua >= 5.1',
}

build = {
    type = 'make',
    build_target = 'all',
    install = {
        lua = {
            ['errors'] = 'errors.lua',
            ['errors.deprecate'] = 'errors/deprecate.lua',
        },
    },
    build_variables = {
        version = '2.1.1-1',
    },
    install_pass = false,
    copy_directories = {'doc'},
}

