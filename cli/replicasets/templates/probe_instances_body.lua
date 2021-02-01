
local cartridge = require('cartridge')

local uris = ...

for _, uri in ipairs(uris) do
    local ok, err = cartridge.admin_probe_server(uri)
    assert(ok, err)
end
