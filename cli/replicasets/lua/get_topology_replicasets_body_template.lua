local cartridge = require('cartridge')

{{ .FormatTopologyReplicasetFunc }}

local topology_replicasets = {}

local replicasets, err = cartridge.admin_get_replicasets()
assert(err == nil, tostring(err))

for _, replicaset in pairs(replicasets) do
    local topology_replicaset = {{ .FormatTopologyReplicasetFuncName }}(replicaset)
    table.insert(topology_replicasets, topology_replicaset)
end

return unpack(topology_replicasets)
