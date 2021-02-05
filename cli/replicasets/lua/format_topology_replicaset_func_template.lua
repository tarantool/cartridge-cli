local function {{ .FormatTopologyReplicasetFuncName }}(replicaset)
    local instances = {}
    for _, server in pairs(replicaset.servers) do
        local instance = {
            alias = server.alias,
            uuid = server.uuid,
            uri = server.uri,
            zone = server.zone,
        }
        table.insert(instances, instance)
    end

    local leader_uuid
    if replicaset.active_master ~= nil then
        leader_uuid = replicaset.active_master.uuid
    end

    local topology_replicaset = {
        uuid = replicaset.uuid,
        alias = replicaset.alias,
        status = replicaset.status,
        roles = replicaset.roles,
        all_rw = replicaset.all_rw,
        weight = replicaset.weight,
        vshard_group = replicaset.vshard_group,
        instances = instances,
        leader_uuid = leader_uuid,
    }

    return topology_replicaset
end
