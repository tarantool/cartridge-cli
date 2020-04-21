local t = require('luatest')
local g = t.group()

local Capture = require('luatest.capture')

local helper = require('test.helper')

local cmd = helper.cartridge_cmd

g.test_version = function()
    local version_str = 'Tarantool cartridge-cli v'
    local capture = Capture:new()
    capture:wrap(true, function() os.execute(cmd .. ' --version') end)
    t.assert_str_contains(capture:flush().stdout, version_str)
    capture:wrap(true, function() os.execute(cmd .. ' -v') end)
    t.assert_str_contains(capture:flush().stdout, version_str)
end
