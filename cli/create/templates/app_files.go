package templates

import "github.com/tarantool/cartridge-cli/cli/templates"

var appFilesTemplate = templates.FileTreeTemplate{
	Dirs: []templates.DirTemplate{
		{
			Path: "app/roles",
			Mode: 0755,
		},
	},
	Files: []templates.FileTemplate{
		{
			Path:    "{{ .Name }}-scm-1.rockspec",
			Mode:    0644,
			Content: rockspecContent,
		},
		{
			Path:    "init.lua",
			Mode:    0755,
			Content: appEntrypointContent,
		},
		{
			Path:    "stateboard.init.lua",
			Mode:    0755,
			Content: stateboardEntrypointContent,
		},
		{
			Path:    "app/roles/custom.lua",
			Mode:    0644,
			Content: customRoleContent,
		},
	},
}

const (
	rockspecContent = `package = '{{ .Name }}'
version = 'scm-1'
source  = {
	url = '/dev/null',
}
-- Put any modules your app depends on here
dependencies = {
	'tarantool',
	'lua >= 5.1',
	'checks == 3.0.1-1',
	'cartridge == 2.1.2-1',
}
build = {
	type = 'none';
}
`

	appEntrypointContent = `#!/usr/bin/env tarantool
require('strict').on()
if package.setsearchroot ~= nil then
	package.setsearchroot()
else
    -- Workaround for rocks loading in tarantool 1.10
    -- It can be removed in tarantool > 2.2
    -- By default, when you do require('mymodule'), tarantool looks into
    -- the current working directory and whatever is specified in
    -- package.path and package.cpath. If you run your app while in the
    -- root directory of that app, everything goes fine, but if you try to
    -- start your app with "tarantool myapp/init.lua", it will fail to load
    -- its modules, and modules from myapp/.rocks.
    local fio = require('fio')
    local app_dir = fio.abspath(fio.dirname(arg[0]))
    print('App dir set to ' .. app_dir)
    package.path = app_dir .. '/?.lua;' .. package.path
    package.path = app_dir .. '/?/init.lua;' .. package.path
    package.path = app_dir .. '/.rocks/share/tarantool/?.lua;' .. package.path
    package.path = app_dir .. '/.rocks/share/tarantool/?/init.lua;' .. package.path
    package.cpath = app_dir .. '/?.so;' .. package.cpath
    package.cpath = app_dir .. '/?.dylib;' .. package.cpath
    package.cpath = app_dir .. '/.rocks/lib/tarantool/?.so;' .. package.cpath
    package.cpath = app_dir .. '/.rocks/lib/tarantool/?.dylib;' .. package.cpath
end
local cartridge = require('cartridge')
local ok, err = cartridge.cfg({
    workdir = 'tmp/db',
    roles = {
        'cartridge.roles.vshard-storage',
        'cartridge.roles.vshard-router',
        'app.roles.custom',
    },
    cluster_cookie = '{{ .Name }}-cluster-cookie',
})
assert(ok, tostring(err))
`

	stateboardEntrypointContent = `#!/usr/bin/env tarantool
require('strict').on()
if package.setsearchroot ~= nil then
    package.setsearchroot()
else
    -- Workaround for rocks loading in tarantool 1.10
    -- It can be removed in tarantool > 2.2
    -- By default, when you do require('mymodule'), tarantool looks into
    -- the current working directory and whatever is specified in
    -- package.path and package.cpath. If you run your app while in the
    -- root directory of that app, everything goes fine, but if you try to
    -- start stateboard with "tarantool myapp/stateboard.init.lua", it will fail to load
    -- its modules, and modules from myapp/.rocks.
    local fio = require('fio')
    local app_dir = fio.abspath(fio.dirname(arg[0]))
    print('App dir set to ' .. app_dir)
    package.path = app_dir .. '/?.lua;' .. package.path
    package.path = app_dir .. '/?/init.lua;' .. package.path
    package.path = app_dir .. '/.rocks/share/tarantool/?.lua;' .. package.path
    package.path = app_dir .. '/.rocks/share/tarantool/?/init.lua;' .. package.path
    package.cpath = app_dir .. '/?.so;' .. package.cpath
    package.cpath = app_dir .. '/?.dylib;' .. package.cpath
    package.cpath = app_dir .. '/.rocks/lib/tarantool/?.so;' .. package.cpath
    package.cpath = app_dir .. '/.rocks/lib/tarantool/?.dylib;' .. package.cpath
end
require('cartridge.stateboard').cfg()
`

	customRoleContent = `local cartridge = require('cartridge')
local function init(opts) -- luacheck: no unused args
    -- if opts.is_master then
    -- end
    local httpd = cartridge.service_get('httpd')
    httpd:route({method = 'GET', path = '/hello'}, function()
        return {body = 'Hello world!'}
    end)
    return true
end
local function stop()
end
local function validate_config(conf_new, conf_old) -- luacheck: no unused args
    return true
end
local function apply_config(conf, opts) -- luacheck: no unused args
    -- if opts.is_master then
    -- end
    return true
end
return {
    role_name = 'app.roles.custom',
    init = init,
    stop = stop,
    validate_config = validate_config,
    apply_config = apply_config,
    -- dependencies = {'cartridge.roles.vshard-router'},
}
`
)
