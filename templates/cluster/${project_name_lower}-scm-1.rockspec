package = '${project_name_lower}'
version = 'scm-1'
source  = {
    url = '/dev/null',
}
-- Put any modules your app depends on here
dependencies = {
    'tarantool',
    'lua >= 5.1',
    'checks == 2.1.1-1',
    'cluster == scm-1',
}
build = {
    type = 'none';
}
