#!/usr/bin/env tarantool

require('strict').on()

if package.setsearchroot ~= nil then
    package.setsearchroot()
end

local log = require('log')
local cluster = require('cluster')
local console = require('console')
local space_explorer = require('space-explorer')

local ok, err = cluster.cfg({
    roles = {
        'cluster.roles.vshard-storage',
        'cluster.roles.vshard-router',
        '${project_name_lower}.custom-role',
    },
})

assert(ok, tostring(err))

--- Add this if you wnat space-explorer - ability to browse raw data in admin UI
space_explorer.init()

if console_sock ~= nil then
    console.listen('unix/:' .. console_sock)
end
