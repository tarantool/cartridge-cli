package templates

import "github.com/tarantool/cartridge-cli/cli/templates"

var testFilesTemplate = templates.FileTreeTemplate{
	Dirs: []templates.DirTemplate{
		{
			Path: "test/helper",
			Mode: 0755,
		},
		{
			Path: "test/integration",
			Mode: 0755,
		},
		{
			Path: "test/unit",
			Mode: 0755,
		},
	},
	Files: []templates.FileTemplate{
		{
			Path:    "test/helper/integration.lua",
			Mode:    0644,
			Content: integrationHelperContent,
		},

		{
			Path:    "test/helper/unit.lua",
			Mode:    0644,
			Content: unitHelperContent,
		},

		{
			Path:    "test/helper.lua",
			Mode:    0644,
			Content: helperContent,
		},

		{
			Path:    "test/integration/api_test.lua",
			Mode:    0644,
			Content: integrationTestContent,
		},

		{
			Path:    "test/unit/sample_test.lua",
			Mode:    0644,
			Content: unitTestContent,
		},
	},
}

const (
	integrationHelperContent = `local t = require('luatest')

local cartridge_helpers = require('cartridge.test-helpers')
local shared = require('test.helper')

local helper = {shared = shared}

helper.cluster = cartridge_helpers.Cluster:new({
    server_command = shared.server_command,
    datadir = shared.datadir,
    use_vshard = false,
    replicasets = {
        {
            alias = 'api',
            uuid = cartridge_helpers.uuid('a'),
            roles = {'app.roles.custom'},
            servers = {
				{ instance_uuid = cartridge_helpers.uuid('a', 1) }
			},
        },
    },
})

t.before_suite(function() helper.cluster:start() end)
t.after_suite(function() helper.cluster:stop() end)

return helper
`

	unitHelperContent = `local t = require('luatest')

local shared = require('test.helper')

local helper = {shared = shared}

t.before_suite(function() box.cfg({work_dir = shared.datadir}) end)

return helper
`

	helperContent = `-- This file is required automatically by luatest.
-- Add common configuration here.

local fio = require('fio')
local t = require('luatest')

local helper = {}

helper.root = fio.dirname(fio.abspath(package.search('init')))
helper.datadir = fio.pathjoin(helper.root, 'tmp', 'db_test')
helper.server_command = fio.pathjoin(helper.root, 'init.lua')

t.before_suite(function()
    fio.rmtree(helper.datadir)
    fio.mktree(helper.datadir)
end)

return helper
`

	integrationTestContent = `local t = require('luatest')
local g = t.group('integration_api')

local helper = require('test.helper.integration')
local cluster = helper.cluster

g.test_sample = function()
    local server = cluster.main_server
    local response = server:http_request('post', '/admin/api', {json = {query = '{}'}})
    t.assert_equals(response.json, {data = {}})
    t.assert_equals(server.net_box:eval('return box.cfg.memtx_dir'), server.workdir)
end
`

	unitTestContent = `local t = require('luatest')
local g = t.group('unit_sample')

require('test.helper.unit')

g.test_sample = function()
    t.assert_equals(type(box.cfg), 'table')
end
`
)
