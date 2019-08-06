local fio = require('fio')
local t = require('luatest')

local shared = require('test.helper')

local helper = {shared = shared}

local root_password = 'secret-password'
helper.server = t.Server:new({
    command = shared.server_command,
    workdir = shared.datadir .. '/3301',
    env = {
        TARANTOOL_ROOT_PASSWORD = root_password,
    },
    net_box_port = 3301,
    net_box_credentials = {
        user = 'root',
        password = root_password,
    }
})

t.before_suite(function()
    fio.mktree(helper.server.workdir)
    helper.server:start()
    t.helpers:retrying(function() helper.server:connect_net_box() end)
end)
t.after_suite(function() helper.server:stop() end)

return helper
