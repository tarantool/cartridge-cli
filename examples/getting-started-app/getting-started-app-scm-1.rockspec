package = 'getting-started-app'
version = 'scm-1'
source  = {
    url = '/dev/null',
}
-- Put any modules your app depends on here
dependencies = {
    'tarantool',
    'lua >= 5.1',
    'checks == 3.1.0-1',
    'cartridge == 2.7.4-1',
    'ldecnumber == 1.1.3-1',
    'metrics == 0.13.0-1',
}
build = {
    type = 'none';
}
