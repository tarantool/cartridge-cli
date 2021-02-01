
local cartridge_roles = require('cartridge.roles')
local known_roles = cartridge_roles.get_known_roles()

local ret = {}
for _, role_name in ipairs(known_roles) do
	local role = {
		name = role_name,
		dependencies = cartridge_roles.get_role_dependencies(role_name),
	}

	table.insert(ret, role)
end

return unpack(ret)
