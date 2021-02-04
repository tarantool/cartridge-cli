local membership = require('membership')

local instances = {}

local members = membership.members()

for uri, member in pairs(members) do
    local uuid
    if member.payload ~= nil and member.payload.uuid ~= nil then
        uuid = member.payload.uuid
    end

    local alias
    if member.payload ~= nil and member.payload.alias ~= nil then
        alias = member.payload.alias
    end

    local instance = {
        uri = uri,
        alias = alias,
        uuid = uuid,
        status = member.status,
    }

    table.insert(instances, instance)
end

return unpack(instances)
