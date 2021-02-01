
local vshard_utils = require('cartridge.vshard-utils')

local known_groups = vshard_utils.get_known_groups()

local known_groups_names = {}
for group_name in pairs(known_groups) do
	table.insert(known_groups_names, group_name)
end

return unpack(known_groups_names)
