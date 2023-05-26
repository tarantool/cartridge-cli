local t = require('luatest')
local g = t.group('unit_sample')

-- create your space here
g.before_all(function(cg) end) -- luacheck: no unused args

-- drop your space here
g.after_all(function(cg) end) -- luacheck: no unused args

g.test_sample = function(cg) -- luacheck: no unused args
    t.assert_equals(type(box.cfg), 'table')
end
