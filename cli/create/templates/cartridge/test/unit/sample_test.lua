local t = require('luatest')
local g = t.group('unit_sample')

g.before_all = function()
    -- create your space here
end

g.after_all = function()
    -- drop your space here
end

g.test_sample = function()
    t.assert_equals(type(box.cfg), 'table')
end
