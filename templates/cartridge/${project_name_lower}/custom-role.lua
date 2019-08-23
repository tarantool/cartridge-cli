#!/usr/bin/env tarantool

local checks = require('checks')
local cartridge = require('cartridge')

local httpd = cartridge.service_registry.get('httpd')
if httpd ~= nil then
    --
end

local function init(opts)
    if opts.is_master then
        --
    end

    --

    return true
end

local function stop()
    --
end

local function validate_config(conf_new, conf_old)
    --

    return true
end

local function apply_config(conf, opts)
    if opts.is_master then
        --
    end

    --

    return true
end

return {
    role_name = 'custom_role',
    init = init,
    stop = stop,
    validate_config = validate_config,
    apply_config = apply_config,
    dependencies = {},
}
