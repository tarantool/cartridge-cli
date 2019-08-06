#!/usr/bin/env tarantool

package.setsearchroot()

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
