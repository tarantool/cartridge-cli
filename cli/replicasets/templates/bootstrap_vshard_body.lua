local cartridge = require('cartridge')

local bootstrap_function = cartridge.admin_bootstrap_vshard
if bootstrap_function == nil then
	bootstrap_function = require('cartridge.admin').bootstrap_vshard
end

local ok, err = bootstrap_function()
assert(ok, tostring(err))
`