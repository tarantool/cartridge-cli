local t = require('luatest')
local g = t.group('integration_api')

local helper = require('test.helper.integration')
local server = helper.server

g.test_sample = function()
    t.assert_equals(server.net_box:eval('return true'), true)
end
