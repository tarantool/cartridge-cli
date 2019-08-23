local t = require('luatest')
local g = t.group('integration_api')

local helper = require('test.helper.integration')
local cluster = helper.cluster

g.test_sample = function()
    t.assertEquals(
        cluster.main_server:http_request('post', '/admin/api', {json = {query = '{}'}}).json,
        {data = {}}
    )
end
