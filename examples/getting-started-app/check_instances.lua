#!/usr/bin/env tarantool

local t = require('luatest')
local yml = require('yaml')
local json = require('json')
local fun = require('fun')
local log = require('log')
local http_client = require('http.client')

local function read_file(path)
    local file = io.open(path, "rb")
    if not file then return nil end
    local content = file:read("*a")
    file:close()
    return content
end

local config = yml.decode(read_file("demo.yml"))


local port = fun.totable(fun.map(
    function(_, obj) return
        obj.http_port
    end,
    config
))[1]
log.info({port=port})

local request_body = {
        operationName = "serverListWithoutStat",
        query =  [[ query serverListWithoutStat {
            serverList: servers {
                uuid
                alias
                uri
                status
                message
                replicaset {
                    uuid
                }
            }
            replicasetList: replicasets {
                alias
                uuid
                status
                roles
                vshard_group
                master {
                    uuid
                }
                active_master {
                    uuid
                }
                weight
                servers {
                    uuid
                    alias
                    uri
                    priority
                    status
                    message
                    replicaset {
                        uuid
                    }
                    labels {
                        name
                        value
                    }
                }
            }
        }]],
        variables = {}
    }

local http_client = http_client.new()
local response = http_client:request('post',
    'http://127.0.0.1:'..tostring(port)..'/admin/api',
    json.encode(request_body)
)

t.assert_equals(response.status, 200)

local ok, json_body = pcall(json.decode, response.body)
if ok then
    response.json = json_body
end

log.info(response.json.data.serverList)

t.assert_items_equals(
    fun.totable(fun.map(
        function(obj) return {uri = obj.uri, alias = obj.alias} end,
        response.json.data.serverList
    )),
    fun.totable(fun.map(
        function(key, obj) return {
            uri = obj.advertise_uri,
            alias = string.gsub(key, "(.*%.)(.*)", "%2")
        } end,
        config
    ))
)