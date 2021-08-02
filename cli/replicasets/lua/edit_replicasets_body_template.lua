local cartridge = require('cartridge')

{{ .FormatTopologyReplicasetFunc }}

local replicasets = ...

local res, err = cartridge.admin_edit_topology({
    replicasets = replicasets,
})

if err ~= nil then
    err = err.err
end

assert(err == nil, tostring(err))

local replicasets = res.replicasets

local topology_replicasets = {}
for _, replicaset in pairs(replicasets) do
    local topology_replicaset = {{ .FormatTopologyReplicasetFuncName }}(replicaset)
    table.insert(topology_replicasets, topology_replicaset)
end

return unpack(topology_replicasets)
