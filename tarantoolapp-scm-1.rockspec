package = 'tarantoolapp'
version = 'scm-1'
source = {
    url = 'git+ssh://git@gitlab.com:tarantool/enterprise/tarantoolapp.git',
    branch = 'master',
}

dependencies = {
    'lua >= 5.1',
}

build = {
    type = 'none',
    copy_directories = {'templates'},

    install = {
        bin = {
            ['tarantoolapp'] = 'tarantoolapp'
        },
    }
}
