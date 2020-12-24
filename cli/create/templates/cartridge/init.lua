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

-- configure cartridge

local cartridge = require('cartridge')

local ok, err = cartridge.cfg({
    roles = {
        'cartridge.roles.vshard-storage',
        'cartridge.roles.vshard-router',
        'cartridge.roles.metrics',
        'app.roles.custom',
    },
    cluster_cookie = '{{ .Name }}-cluster-cookie',
})

assert(ok, tostring(err))

-- register admin function probe to use it with "cartridge admin"

local cli_admin = require('cartridge-cli-extensions.admin')

cli_admin.init()

local probe = {
    usage = 'Probe instance',
    args = {
        uri = {
            type = 'string',
            usage = 'Instance URI',
        },
    },
    call = function(opts)
        opts = opts or {}

        if opts.uri == nil then
            return nil, "Please, pass instance URI via --uri flag"
        end

        local cartridge_admin = require('cartridge.admin')
        local ok, err = cartridge_admin.probe_server(opts.uri)

        if not ok then
            return nil, err.err
        end

        return {
            string.format('Probe %q: OK', opts.uri),
        }
    end,
}

local ok, err = cli_admin.register('probe', probe.usage, probe.args, probe.call)
assert(ok, err)
