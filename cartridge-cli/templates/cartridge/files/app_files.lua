local app_files = {
    {
        name = '${project_name_lower}-scm-1.rockspec',
        mode = tonumber('0644', 8),
        content = [=[
            package = '${project_name_lower}'
            version = 'scm-1'
            source  = {
                url = '/dev/null',
            }
            -- Put any modules your app depends on here
            dependencies = {
                'tarantool',
                'lua >= 5.1',
                'checks == 3.0.1-1',
                'cartridge == scm-1',
            }
            build = {
                type = 'none';
            }
        ]=]
    },
    {
        name = 'app/roles/custom.lua',
        mode = tonumber('0644', 8),
        content = [=[
            local cartridge = require('cartridge')

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
        ]=]
    },
    {
        name = 'init.lua',
        mode = tonumber('0755', 8),
        content = [=[
            #!/usr/bin/env tarantool

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
                cluster_cookie = '${project_name_lower}-cluster-cookie',
            })

            assert(ok, tostring(err))

        ]=]
    },
    {
        name = 'stateboard.init.lua',
        mode = tonumber('0755', 8),
        content = [=[
            #!/usr/bin/env tarantool

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
        ]=]
    },
}

return app_files
