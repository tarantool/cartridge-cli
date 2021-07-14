local cartridge = require('cartridge')
local res, err = require('cartridge').failover_set_params(...)

if err ~= nil then
    return nil, err.err
end

return res, nil
