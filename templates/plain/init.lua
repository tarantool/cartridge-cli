#!/usr/bin/env tarantool

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
    package.path = package.path .. ';' .. app_dir .. '/?.lua'
    package.path = package.path .. ';' .. app_dir .. '/?/init.lua'
    package.path = package.path .. ';' .. app_dir .. '/.rocks/share/tarantool/?.lua'
    package.path = package.path .. ';' .. app_dir .. '/.rocks/share/tarantool/?/init.lua'
    package.cpath = package.cpath .. ';' .. app_dir .. '/?.so'
    package.cpath = package.cpath .. ';' .. app_dir .. '/?.dylib'
    package.cpath = package.cpath .. ';' .. app_dir .. '/.rocks/lib/tarantool/?.so'
    package.cpath = package.cpath .. ';' .. app_dir .. '/.rocks/lib/tarantool/?.dylib'
end

local workdir = os.getenv('TARANTOOL_WORKDIR') or 'tmp/db'
require('fio').mktree(workdir)

-- When starting multiple instances of the app from systemd,
-- instance_name will contain the part after the "@". e.g.  for
-- myapp@instance_1, instance_name will contain "instance_1".
-- Then we use the suffix to assign port number, so that
-- listen port will be base_listen + suffix
local instance_name = os.getenv('TARANTOOL_INSTANCE_NAME')
local instance_id = instance_name and tonumber(string.match(instance_name, "_(%d+)$"))

local listen
if instance_id then
    print("Instance name: " .. instance_name)

    local base_listen = os.getenv('TARANTOOL_BASE_LISTEN') or 3300
    listen = base_listen + instance_id
else
    listen = os.getenv('TARANTOOL_LISTEN') or 3301
end

box.cfg({
    work_dir = workdir,
    listen = listen,
})

local root_password = os.getenv('TARANTOOL_ROOT_PASSWORD')
if root_password and root_password:len() > 0 then
    box.schema.user.create('root', {password = root_password, if_not_exists = true})
    box.schema.user.grant('root', 'read,write,execute', 'universe')
end

local console_sock = os.getenv('TARANTOOL_CONSOLE_SOCK')
if console_sock ~= nil then
    local console = require('console')
    console.listen('unix/:' .. console_sock)
end
