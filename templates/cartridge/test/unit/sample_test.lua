local t = require('luatest')
local g = t.group('unit_sample')

require('test.helper.unit')

g.test_sample = function()
    t.assertEquals(type(box.cfg), 'table')
end
