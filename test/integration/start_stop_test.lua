local t = require('luatest')
local g = t.group()

local fio = require('fio')

local Capture = require('luatest.capture')

local helper = require('test.helper')
local cmd = helper.cartridge_cmd

local RUN_DIR = fio.pathjoin(helper.tempdir, 'test_run')
local INSTANCE_SCRIPT = 'test/instances/init.lua'
local TEST_OPTS = {'--run-dir', RUN_DIR}
local SIMPLE_INSTANCE_OPTS = helper.concat({'--script',INSTANCE_SCRIPT }, TEST_OPTS)
local INSTANCE_PIDFILE = fio.pathjoin(RUN_DIR, 'cartridge-cli.test_name.pid')

g.before_each(function() fio.rmtree(RUN_DIR) end)

g.test_start_stop = function()
    local starter = helper.os_execute(cmd, helper.concat({'start', '.test_name', '-d'}, SIMPLE_INSTANCE_OPTS))
    local pid = tonumber(helper.read_file(INSTANCE_PIDFILE))
    t.assert_not_equals(pid, starter.pid)
    t.assert(helper.check_pid_running(pid))
    helper.os_execute(cmd, helper.concat({'stop', '.test_name'}, TEST_OPTS))
    t.assert_not(helper.check_pid_running(pid))
    t.assert_not(fio.stat(INSTANCE_PIDFILE))
end

g.test_start_stop_with_options_in_env = function()
    local starter = helper.os_execute(cmd, {'start', '.test_name', '-d'}, {
        env = {
            TARANTOOL_SCRIPT = INSTANCE_SCRIPT,
            TARANTOOL_RUN_DIR = RUN_DIR,
        }
    })
    local pid = tonumber(helper.read_file(INSTANCE_PIDFILE))
    t.assert_not_equals(pid, starter.pid)
    t.assert(helper.check_pid_running(pid))
    helper.os_execute(cmd, {'stop', '.test_name'}, {env = {TARANTOOL_RUN_DIR = RUN_DIR}})
    t.assert_not(helper.check_pid_running(pid))
end

local function assert_start_stop_all(config_opts, instance_names)
    local starter = helper.os_execute(cmd, helper.concat({'start', '-d'}, config_opts, SIMPLE_INSTANCE_OPTS), {
        timeout = 5,
    })
    instance_names = instance_names or
        {'test_app.storage_1', 'test_app.storage_2', 'test_app.router_1'}
    local pids_by_instance_name = {}
    for _, instance_name in pairs(instance_names) do
        local pid = tonumber(helper.read_file('tmp/test_run/' .. instance_name .. '.pid'))
        t.assert_not_equals(pid, starter.pid)
        t.assert(helper.check_pid_running(pid))
        pids_by_instance_name[instance_name] = pid
    end
    helper.os_execute(cmd, helper.concat({'stop'}, config_opts, TEST_OPTS))
    for _, instance_name in pairs(instance_names) do
        t.assert_not(fio.stat('tmp/test_run/' .. instance_name .. '.pid'))
        t.assert_not(helper.check_pid_running(pids_by_instance_name[instance_name]))
    end
end

g.test_start_stop_all = function()
    assert_start_stop_all({'test_app', '--cfg', 'test/instances/instances.yml'})
end

g.test_start_stop_all_with_split_config = function()
    assert_start_stop_all({'test_app', '--cfg', 'test/instances/config_multiple'})
end

g.test_start_stop_all_with_app_name_from_rockspec = function()
    assert_start_stop_all(
        {'--cfg', 'test/instances/instances.yml'},
        {'cartridge-cli.cli_instance_1', 'cartridge-cli.cli_instance_2'}
    )
end

g.test_start_stop_all_with_invalid_app_name = function()
    local capture = Capture:new()
    capture:wrap(true, function()
        helper.os_execute(cmd, helper.concat(
            {'start', 'tdg', '--cfg', 'test/instances/config_multiple', '-d'}, SIMPLE_INSTANCE_OPTS
        ))
    end)
    t.assert_str_contains(capture:flush().stderr, 'No configured instances found for app tdg')
end

g.test_start_stop_all_with_apps_path = function()
    assert_start_stop_all(
        {'instances', '--cfg', 'test/instances/instances.yml', '--apps-path', fio.abspath('test')},
        {'instances.app_path_1', 'instances.app_path_2'}
    )
end

g.test_start_with_missed_entrypoint_script = function()
    local capture = Capture:new()
    capture:wrap(true, function()
        helper.os_execute(cmd, helper.concat(
            {'start', 'test_app'},
            {'--cfg', 'test/instances/instances.yml'},
            {'--script', 'non-existent-script.lua'},
            TEST_OPTS
        ))
    end)
    t.assert_str_contains(capture:flush().stderr, 'Application entrypoint script does not exists')
end
