package = '{{ .Name }}'
version = 'scm-1'
source  = {
	url = '/dev/null',
}
-- Put any modules your app depends on here
dependencies = {
    'tarantool',
    'lua >= 5.1',
    'checks == 3.0.1-1',
    'cartridge == 2.3.0-1',
    'metrics == 0.5.0-1',
    'cartridge-cli-extensions == 1.1.0-1',
}
build = {
	type = 'none';
}
