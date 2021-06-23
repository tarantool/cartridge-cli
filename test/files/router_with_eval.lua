local cartridge = require('cartridge')

local function eval_handler(req)
    local resp = req:render({json = { data = loadstring(req:json().eval_string)() }})
    return resp
end

local function init(opts) -- luacheck: no unused args
    local httpd = assert(cartridge.service_get('httpd'), "Failed to get httpd serivce")
    httpd:route({method = 'PUT', path = '/eval'}, eval_handler)

    return true
end

return {
    role_name = 'app.roles.custom',
    init = init,
}
