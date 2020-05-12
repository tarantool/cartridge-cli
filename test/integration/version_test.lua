local t = require('luatest')
local g = t.group()

local Capture = require('luatest.capture')

local helper = require('test.helper')

local cmd = helper.cartridge_cmd

g.test_version = function()
    t.skip()

    local version_str = 'Cartridge CLI v'
    local capture = Capture:new()
    capture:wrap(true, function() os.execute(cmd .. ' version') end)
    t.assert_str_contains(capture:flush().stdout, version_str)
end
