#!/usr/bin/env tarantool

package.setsearchroot()


local work_dir = os.getenv('TARANTOOL_WORK_DIR') or '.'

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
    work_dir = work_dir,
    listen = listen,
})

local console_sock = os.getenv('TARANTOOL_CONSOLE_SOCK')
if console_sock ~= nil then
    local console = require('console')
    console.listen('unix/:' .. console_sock)
end
