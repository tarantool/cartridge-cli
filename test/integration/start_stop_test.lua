local t = require('luatest')
local g = t.group()

local fio = require('fio')

local Capture = require('luatest.capture')

local helper = require('test.helper')
local cmd = helper.cartridge_cmd

local RUN_DIR = fio.pathjoin(helper.tempdir, 'test_run')
local INSTANCE_SCRIPT = 'test/instances/init.lua'
local IGNORE_SIGTERM_SCRIPT = 'test/instances/init.ignore_sigterm.lua'
local TEST_OPTS = {'--run-dir', RUN_DIR}
local SIMPLE_INSTANCE_OPTS = helper.concat({'--script',INSTANCE_SCRIPT }, TEST_OPTS)
local INSTANCE_PIDFILE = fio.pathjoin(RUN_DIR, 'cartridge-cli.test_name.pid')

g.before_each(function() fio.rmtree(RUN_DIR) end)

g.test_start_stop = function()
    t.skip()

    local starter = helper.os_execute(cmd, helper.concat({'start', '.test_name', '-d'}, SIMPLE_INSTANCE_OPTS))
    local pid = tonumber(helper.read_file(INSTANCE_PIDFILE))
    t.assert_not_equals(pid, starter.pid)
    t.assert(helper.check_pid_running(pid))
    helper.os_execute(cmd, helper.concat({'stop', '.test_name'}, TEST_OPTS))
    t.assert_not(helper.check_pid_running(pid))
end

g.test_start_stop_with_options_in_env = function()
    t.skip()

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
        t.assert_not(helper.check_pid_running(pids_by_instance_name[instance_name]))
    end
end

g.test_start_stop_all = function()
    t.skip()

    assert_start_stop_all({'test_app', '--cfg', 'test/instances/instances.yml'})
end

g.test_start_stop_all_with_split_config = function()
    t.skip()

    assert_start_stop_all({'test_app', '--cfg', 'test/instances/config_multiple'})
end

g.test_start_stop_all_with_app_name_from_rockspec = function()
    t.skip()

    assert_start_stop_all(
        {'--cfg', 'test/instances/instances.yml'},
        {'cartridge-cli.i1', 'cartridge-cli.i2'}
    )
end

g.test_start_stop_all_with_invalid_app_name = function()
    t.skip()

    local capture = Capture:new()
    capture:wrap(true, function()
        helper.os_execute(cmd, helper.concat(
            {'start', 'tdg', '--cfg', 'test/instances/config_multiple', '-d'}, SIMPLE_INSTANCE_OPTS
        ))
    end)
    t.assert_str_contains(capture:flush().stderr, 'No configured instances found for app tdg')
end

g.test_start_stop_all_with_apps_path = function()
    t.skip()

    assert_start_stop_all(
        {'instances', '--cfg', 'test/instances/instances.yml', '--apps-path', fio.abspath('test')},
        {'instances.app_path_1', 'instances.app_path_2'}
    )
end

g.test_start_with_missed_entrypoint_script = function()
    t.skip()

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

g.test_notify_socket_length = function()
    t.skip()

    local long_run_dir = fio.pathjoin(helper.tempdir, string.rep('a', 110))
    fio.mktree(long_run_dir)

    local capture = Capture:new()
    capture:wrap(true, function()
        helper.os_execute(cmd, helper.concat(
            {'start', '.test_name', '-d'},
            {'--cfg', 'test/instances/instances.yml'},
            {'--script',INSTANCE_SCRIPT},
            {'--run-dir', long_run_dir}
        ))
    end)
    t.assert_str_contains(capture:flush().stderr, 'Too long notify socket name exceeds UNIX_PATH_MAX limit')
end

g.test_sigterm_ignored = function()
    t.skip()

    local CARTRIGDE_STOP_TIMEOUT = 1

    local starter = helper.os_execute(cmd,
        helper.concat(
            {'start', '.test_name', '-d'},
            TEST_OPTS,
            {'--script', IGNORE_SIGTERM_SCRIPT}
        )
    )
    local pid = tonumber(helper.read_file(INSTANCE_PIDFILE))
    t.assert_not_equals(pid, starter.pid)
    t.assert(helper.check_pid_running(pid))

    local capture = Capture:new()
    capture:wrap(true, function()
        helper.os_execute(cmd,
                helper.concat(
                    {'stop', '.test_name'},
                    TEST_OPTS
                ),
                {
                    env = {CARTRIGDE_STOP_TIMEOUT = CARTRIGDE_STOP_TIMEOUT},
                    timeout = CARTRIGDE_STOP_TIMEOUT + 1
                }
            )
    end)
    t.assert_str_contains(
        capture:flush().stderr,
        string.format('Can not kill process %s: it is still running', pid)
    )

    t.assert(helper.check_pid_running(pid))
    t.assert(fio.path.exists(INSTANCE_PIDFILE))

    helper.kill_process(pid)
end
