#!/usr/bin/env tarantool

local t = require('luatest')
local yml = require('yaml')
local json = require('json')
local fun = require('fun')
local log = require('log')
local http_client_lib = require('http.client')

local function read_file(path)
    local file = io.open(path, "rb")
    if not file then return nil end
    local content = file:read("*a")
    file:close()
    return content
end

local config = yml.decode(read_file("demo.yml"))

local http_port = fun.iter(config)
        :map(function(_, obj) return obj.http_port end)
        :totable()[1]

log.info({http_port=http_port})

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

local http_client = http_client_lib.new()
t.helpers.retrying(
    {timeout = 3, delay = 0.2},
    function()
        local response = http_client:request('post',
            'http://127.0.0.1:'..tostring(http_port)..'/admin/api',
            json.encode(request_body)
        )

        local ok, json_body = pcall(json.decode, response.body)
        if ok then
            response.json = json_body
        else
            log.info({status=response.status})
            t.assert(false)
        end

        local actual = fun.iter(response.json.data.serverList)
                        :map(function(obj) return {uri = obj.uri, alias = obj.alias} end)
                        :totable()

        local expected = fun.iter(config)
                        :map(function(key, obj)
                                return {
                                    uri = obj.advertise_uri,
                                    alias = string.gsub(key, "(.*%.)(.*)", "%2")
                                }
                            end
                        )
                        :totable()

        log.info({actual=actual})
        log.info({expected=expected})

        log.info("retry...")

        t.assert_items_equals(actual, expected)
    end
)