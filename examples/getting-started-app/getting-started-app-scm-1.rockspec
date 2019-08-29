package = 'getting-started-app'
version = 'scm-1'
source  = {
    url = '/dev/null',
}
-- Put any modules your app depends on here
dependencies = {
    'tarantool',
    'lua >= 5.1',
    'luatest == 0.2.0-1',
    'ldecnumber == 1.1.3-1',
    'cartridge == scm-1',
}
build = {
    type = 'none';
}