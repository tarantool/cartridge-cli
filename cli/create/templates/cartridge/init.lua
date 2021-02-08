#!/usr/bin/env tarantool

require('strict').on()

-- configure path

local path = require('app.path')
local config = path.cfg()

package.path = config.path
package.cpath = config.cpath
package.setsearchroot(config.root)

-- configure cartridge

local cartridge = require('cartridge')

local ok, err = cartridge.cfg({
    roles = {
        'cartridge.roles.vshard-storage',
        'cartridge.roles.vshard-router',
        'cartridge.roles.metrics',
        'app.roles.custom',
    },
    cluster_cookie = '{{ .Name }}-cluster-cookie',
})

assert(ok, tostring(err))

-- register admin function to use it with 'cartridge admin' command

local admin = require('app.admin')
admin.init()

local metrics = require('cartridge.roles.metrics')
metrics.set_export({
    {
        path = '/metrics',
        format = 'prometheus'
    },
    {
        path = '/health',
        format = 'health'
    }
})

