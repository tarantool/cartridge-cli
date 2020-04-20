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
    local instance_pidfile = fio.pathjoin(RUN_DIR, string.format('%s.pid', instance_fullname))
    print(instance_pidfile)
    t.assert(fio.path.exists(instance_pidfile))

    local instance_pid = tonumber(helper.read_file(instance_pidfile))
    t.assert(helper.check_pid_running(instance_pid))

    running_pids[instance_fullname] = instance_pid
end

local function check_is_not_running(instance_fullname)
    local instance_pidfile = fio.pathjoin(RUN_DIR, string.format('%s.pid', instance_fullname))
    t.assert_not(fio.path.exists(instance_pidfile))

    local instance_pid = running_pids[instance_fullname]
    if instance_pid ~= nil then
        t.assert_not(helper.check_pid_running(instance_pid))
    end
end

g.test_one_instance_specified = function()
    local INSTANCE_FULLNAME = string.format('%s.storage_1', TEST_APP_NAME)
    local STATEBOARD_FULLNAME = string.format('%s-stateboard', TEST_APP_NAME)

    local capture = Capture:new()
    capture:wrap(true, function()
        helper.os_execute(cmd,
            helper.concat(
                {'start', '-d', INSTANCE_FULLNAME},
                RUN_DIR_OPT,
                {'--stateboard-only'}
            ),
            {chdir = TEST_APP_DIR}
        )
    end)

    t.assert_str_contains(
        capture:flush().stderr,
        string.format('Passed instance ID (%s) is ignored', INSTANCE_FULLNAME)
    )

    check_is_not_running(INSTANCE_FULLNAME)
    check_is_running(STATEBOARD_FULLNAME)

    -- stop with --stateboard
    local capture = Capture:new()
    capture:wrap(true, function()
        helper.os_execute(cmd,
            helper.concat(
                {'stop', INSTANCE_FULLNAME},
                RUN_DIR_OPT,
                {'--stateboard-only'}
            ),
            {chdir = TEST_APP_DIR}
        )
    end)

    t.assert_str_contains(
        capture:flush().stderr,
        string.format('Passed instance ID (%s) is ignored', INSTANCE_FULLNAME)
    )

    check_is_not_running(STATEBOARD_FULLNAME)
end

g.test_instances_described_in_config = function()
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
            {'--stateboard-only'}
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
            {'--stateboard-only'}
        ),
        {chdir = TEST_APP_DIR}
    )

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
                {'--stateboard-only'}
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
    helper.write_file(CLI_CONF_PATH, 'stateboard_only: true\n')

    -- start
    helper.os_execute(cmd,
        helper.concat(
            {'start', '-d', INSTANCE_FULLNAME},
            RUN_DIR_OPT
        ),
        {chdir = TEST_APP_DIR}
    )

    check_is_not_running(INSTANCE_FULLNAME)
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
