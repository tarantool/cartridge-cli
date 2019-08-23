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
    type = 'none',
    copy_directories = {'templates'},

    install = {
        bin = {
            ['cartridge'] = 'cartridge'
        },
    }
}
