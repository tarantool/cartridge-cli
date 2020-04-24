local t = require('luatest')
local g = t.group()

local fio = require('fio')

local Capture = require('luatest.capture')

local helper = require('test.helper')
local cmd = helper.cartridge_cmd

local TEST_APP_NAME = 'test_app'

local RUN_DIR = fio.pathjoin(helper.tempdir, 'test_run')
local TEST_APP_DIR = fio.pathjoin(helper.tempdir, TEST_APP_NAME)

local RUN_DIR_OPT = {'--run-dir', RUN_DIR}

local running_pids = {}

g.before_each(function()
    fio.rmtree(RUN_DIR)
    fio.rmtree(TEST_APP_DIR)
    fio.copytree(
        fio.pathjoin(fio.cwd(), 'test/integration/test_app'),
        TEST_APP_DIR
    )
    running_pids = {}
end)

g.after_each(function()
    for _, pid in pairs(running_pids) do
        helper.kill_process(pid)
    end
end)

local function check_is_running(instance_fullname)
    helper.check_is_running(instance_fullname, running_pids, RUN_DIR)
end

local function check_is_not_running(instance_fullname)
    helper.check_is_not_running(instance_fullname, running_pids)
end

g.test_one_instance = function()
    local INSTANCE_FULLNAME = string.format('%s.storage_1', TEST_APP_NAME)
    local STATEBOARD_FULLNAME = string.format('%s-stateboard', TEST_APP_NAME)

    helper.os_execute(cmd,
        helper.concat(
            {'start', '-d', INSTANCE_FULLNAME},
            RUN_DIR_OPT,
            {'--stateboard'}
        ),
        {chdir = TEST_APP_DIR}
    )

    check_is_running(INSTANCE_FULLNAME)
    check_is_running(STATEBOARD_FULLNAME)

    -- stop without --stateboard
    helper.os_execute(cmd,
        helper.concat(
            {'stop', INSTANCE_FULLNAME},
            RUN_DIR_OPT
        ),
        {chdir = TEST_APP_DIR}
    )
    check_is_not_running(INSTANCE_FULLNAME)
    check_is_running(STATEBOARD_FULLNAME)

    -- stop with --stateboard
    helper.os_execute(cmd,
        helper.concat(
            {'stop', INSTANCE_FULLNAME},
            RUN_DIR_OPT,
            {'--stateboard'}
        ),
        {chdir = TEST_APP_DIR}
    )

    check_is_not_running(INSTANCE_FULLNAME)
    check_is_not_running(STATEBOARD_FULLNAME)
end

g.test_instances_from_config = function()
    local INSTANCE_1_FULLNAME = string.format('%s.instance-1', TEST_APP_NAME)
    local INSTANCE_2_FULLNAME = string.format('%s.instance-2', TEST_APP_NAME)
    local STATEBOARD_FULLNAME = string.format('%s-stateboard', TEST_APP_NAME)

    local CFG_PATH = fio.pathjoin(TEST_APP_DIR, 'instances.yml')

    local instances_cfg = table.concat({
       string.format('%s:', INSTANCE_1_FULLNAME),
       string.format('%s:', INSTANCE_2_FULLNAME)
    }, '\n')
    helper.write_file(CFG_PATH, instances_cfg)

    helper.os_execute(cmd,
        helper.concat(
            {'start', '-d', TEST_APP_NAME},
            RUN_DIR_OPT,
            {'--stateboard'}
        ),
        {chdir = TEST_APP_DIR}
    )

    check_is_running(INSTANCE_1_FULLNAME)
    check_is_running(INSTANCE_2_FULLNAME)
    check_is_running(STATEBOARD_FULLNAME)

    -- stop without --stateboard
    helper.os_execute(cmd,
        helper.concat(
            {'stop', TEST_APP_NAME},
            RUN_DIR_OPT
        ),
        {chdir = TEST_APP_DIR}
    )
    check_is_not_running(INSTANCE_1_FULLNAME)
    check_is_not_running(INSTANCE_2_FULLNAME)
    check_is_running(STATEBOARD_FULLNAME)

    -- stop with --stateboard
    helper.os_execute(cmd,
        helper.concat(
            {'stop', TEST_APP_NAME},
            RUN_DIR_OPT,
            {'--stateboard'}
        ),
        {chdir = TEST_APP_DIR}
    )

    check_is_not_running(INSTANCE_1_FULLNAME)
    check_is_not_running(INSTANCE_2_FULLNAME)
    check_is_not_running(STATEBOARD_FULLNAME)
end

g.test_start_with_non_existent_stateboard_entrypoint = function()
    local INSTANCE_FULLNAME = string.format('%s.storage_1', TEST_APP_NAME)
    local STATEBOARD_FULLNAME = string.format('%s-stateboard', TEST_APP_NAME)

    fio.unlink(fio.pathjoin(TEST_APP_DIR, 'stateboard.init.lua'))

    local capture = Capture:new()
    capture:wrap(true, function()
        helper.os_execute(cmd,
        helper.concat(
                {'start', '-d', INSTANCE_FULLNAME},
                RUN_DIR_OPT,
                {'--stateboard'}
            ),
            {chdir = TEST_APP_DIR}
        )
    end)
    t.assert_str_contains(capture:flush().stderr, 'Stateboard entrypoint script does not exists')

    check_is_not_running(INSTANCE_FULLNAME)
    check_is_not_running(STATEBOARD_FULLNAME)
end

g.test_flag_from_conf = function()
    local INSTANCE_FULLNAME = string.format('%s.storage_1', TEST_APP_NAME)
    local STATEBOARD_FULLNAME = string.format('%s-stateboard', TEST_APP_NAME)

    local CLI_CONF_PATH = fio.pathjoin(TEST_APP_DIR, '.cartridge.yml')
    helper.write_file(CLI_CONF_PATH, 'stateboard: true\n')

    -- start
    helper.os_execute(cmd,
        helper.concat(
            {'start', '-d', INSTANCE_FULLNAME},
            RUN_DIR_OPT
        ),
        {chdir = TEST_APP_DIR}
    )

    check_is_running(INSTANCE_FULLNAME)
    check_is_running(STATEBOARD_FULLNAME)

    -- stop
    helper.os_execute(cmd,
        helper.concat(
            {'stop', INSTANCE_FULLNAME},
            RUN_DIR_OPT
        ),
        {chdir = TEST_APP_DIR}
    )

    check_is_not_running(INSTANCE_FULLNAME)
    check_is_not_running(STATEBOARD_FULLNAME)
end

g.test_flag_from_env = function()
    local INSTANCE_FULLNAME = string.format('%s.storage_1', TEST_APP_NAME)
    local STATEBOARD_FULLNAME = string.format('%s-stateboard', TEST_APP_NAME)

    -- start
    helper.os_execute(cmd,
        helper.concat(
            {'start', '-d', INSTANCE_FULLNAME},
            RUN_DIR_OPT
        ),
        {
            chdir = TEST_APP_DIR,
            env = {TARANTOOL_STATEBOARD = 'true'},
        }
    )

    check_is_running(INSTANCE_FULLNAME)
    check_is_running(STATEBOARD_FULLNAME)

    -- stop
    helper.os_execute(cmd,
        helper.concat(
            {'stop', INSTANCE_FULLNAME},
            RUN_DIR_OPT
        ),
        {
            chdir = TEST_APP_DIR,
            env = {TARANTOOL_STATEBOARD = 'true'},
        }
    )

    check_is_not_running(INSTANCE_FULLNAME)
    check_is_not_running(STATEBOARD_FULLNAME)

    -- test passing wrong value in a flag
    local capture = Capture:new()
    capture:wrap(true, function()
        helper.os_execute(cmd,
            helper.concat(
                {'stop', INSTANCE_FULLNAME},
                RUN_DIR_OPT
            ),
            {
                chdir = TEST_APP_DIR,
                env = {TARANTOOL_STATEBOARD = 'wrong value'},
            }
        )
    end)
    t.assert_str_contains(
        capture:flush().stderr,
        'Cannot get TARANTOOL_STATEBOARD from env: value should be `true` or `false`'
    )

    check_is_not_running(INSTANCE_FULLNAME)
    check_is_not_running(STATEBOARD_FULLNAME)
end
