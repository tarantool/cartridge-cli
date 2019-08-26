package = '${project_name_lower}'
version = 'scm-1'
source  = {
    url = '/dev/null',
}
-- Put any modules your app depends on here
dependencies = {
    'tarantool',
    'lua >= 5.1',
    'luatest == 0.2.0-1',
    'cartridge == 1.0.0-1',
}
build = {
    type = 'none';
}
