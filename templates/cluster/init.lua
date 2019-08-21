#!/usr/bin/env tarantool

require('strict').on()

package.setsearchroot()

local log = require('log')
local cluster = require('cluster')
local console = require('console')
local space_explorer = require('space-explorer')

local work_dir = os.getenv("TARANTOOL_WORK_DIR") or '.'
local instance_name = os.getenv("TARANTOOL_INSTANCE_NAME")
local console_sock = os.getenv("TARANTOOL_CONSOLE_SOCK")
local memtx_memory = os.getenv("TARANTOOL_MEMTX_MEMORY")

local base_advertise_port = os.getenv("TARANTOOL_BASE_ADVERTISE_PORT") or 3300
local base_http_port = os.getenv("TARANTOOL_BASE_HTTP_PORT") or 8080

local advertise_port = os.getenv("TARANTOOL_ADVERTISE_PORT") or 3301
local http_port = os.getenv("TARANTOOL_HTTP_PORT") or 8081
local hostname = os.getenv("TARANTOOL_HOSTNAME") or "localhost"

local bucket_count = os.getenv("TARANTOOL_BUCKET_COUNT") or 30000

-- When starting multiple instances of the app from systemd,
-- instance_name will contain the part after the "@". e.g.  for
-- myapp@instance_1, instance_name will contain "instance_1".
-- Then we use the suffix to assign port number, so that
-- advertise port will be base_adevrtise_port + suffix
if instance_name ~= nil then
    print("Instance name: " .. instance_name)

    local instance_no = string.match(instance_name, "_(%d+)$")
    if instance_no ~= nil then
        advertise_port = base_advertise_port + tonumber(instance_no)
        http_port = base_http_port + tonumber(instance_no)
    end
end

local ok, err = cluster.cfg({ -- Enterprise cluster-specific parameters
    alias = instance_name,
    workdir = work_dir,
    advertise_uri = string.format('%s:%s', hostname, advertise_port),
    cluster_cookie = 'secret-cluster-cookie',
    bucket_count = bucket_count,
    http_port = http_port,
    roles = {
        'cluster.roles.vshard-storage',
        'cluster.roles.vshard-router',
        '${project_name_lower}.custom-role',
    },
}, { -- additional box.cfg parameters that require tuning
    memtx_memory = memtx_memory
})

assert(ok, tostring(err))

--- Add this if you wnat space-explorer - ability to browse raw data in admin UI
space_explorer.init()

if console_sock ~= nil then
    console.listen('unix/:' .. console_sock)
end
