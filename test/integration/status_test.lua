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

local function get_status_strings(instance_id, opts)
    local capture = Capture:new()
    capture:wrap(true, function()
        helper.os_execute(cmd,
            helper.concat(
                {'status', instance_id},
                RUN_DIR_OPT,
                opts or {}
            ),
            {chdir = TEST_APP_DIR}
        )
    end)

    local output = capture:flush().stderr
    local output_strings = output:strip():split('\n')
    table.sort(output_strings)
    return output_strings
end

local function start(instance_id, opts)
    helper.os_execute(cmd,
        helper.concat(
            {'start', '-d', instance_id},
            RUN_DIR_OPT,
            opts or {}
        ),
        {chdir = TEST_APP_DIR}
    )
end

local function stop(instance_id, opts)
    helper.os_execute(cmd,
        helper.concat(
            {'stop', instance_id},
            RUN_DIR_OPT,
            opts or {}
        ),
        {chdir = TEST_APP_DIR}
    )
end

g.test_one_instance = function()
    t.skip()

    local INSTANCE_FULLNAME = string.format('%s.storage_1', TEST_APP_NAME)
    local STATEBOARD_FULLNAME = string.format('%s-stateboard', TEST_APP_NAME)

    -- instance - not started
    -- stateboard - not started
    check_is_not_running(INSTANCE_FULLNAME)
    check_is_not_running(STATEBOARD_FULLNAME)

    local status_strings = get_status_strings(INSTANCE_FULLNAME)
    t.assert_equals(#status_strings, 1)
    t.assert_str_contains(status_strings[1], INSTANCE_FULLNAME)
    t.assert_str_contains(status_strings[1], 'Not started')

    local status_strings = get_status_strings(TEST_APP_NAME, {'--stateboard-only'})
    t.assert_equals(#status_strings, 1)
    t.assert_str_contains(status_strings[1], STATEBOARD_FULLNAME)
    t.assert_str_contains(status_strings[1], 'Not started')

    -- start instance
    start(INSTANCE_FULLNAME)
    check_is_running(INSTANCE_FULLNAME)

    -- instance - running
    -- stateboard - not started
    local status_strings = get_status_strings(INSTANCE_FULLNAME)
    t.assert_equals(#status_strings, 1)
    t.assert_str_contains(status_strings[1], INSTANCE_FULLNAME)
    t.assert_str_contains(status_strings[1], 'Running')

    local status_strings = get_status_strings(TEST_APP_NAME, {'--stateboard-only'})
    t.assert_equals(#status_strings, 1)
    t.assert_str_contains(status_strings[1], STATEBOARD_FULLNAME)
    t.assert_str_contains(status_strings[1], 'Not started')

    local status_strings = get_status_strings(INSTANCE_FULLNAME, {'--stateboard'})
    t.assert_equals(#status_strings, 2)
    t.assert_str_contains(status_strings[1], STATEBOARD_FULLNAME)
    t.assert_str_contains(status_strings[1], 'Not started')
    t.assert_str_contains(status_strings[2], INSTANCE_FULLNAME)
    t.assert_str_contains(status_strings[2], 'Running')

    -- stop instance
    stop(INSTANCE_FULLNAME)
    check_is_not_running(INSTANCE_FULLNAME)

    -- instance - stopped
    -- stateboard - not started
    local status_strings = get_status_strings(INSTANCE_FULLNAME)
    t.assert_equals(#status_strings, 1)
    t.assert_str_contains(status_strings[1], INSTANCE_FULLNAME)
    t.assert_str_contains(status_strings[1], 'Stopped')

    local status_strings = get_status_strings(INSTANCE_FULLNAME, {'--stateboard'})
    t.assert_equals(#status_strings, 2)
    t.assert_str_contains(status_strings[1], STATEBOARD_FULLNAME)
    t.assert_str_contains(status_strings[1], 'Not started')
    t.assert_str_contains(status_strings[2], INSTANCE_FULLNAME)
    t.assert_str_contains(status_strings[2], 'Stopped')
end

g.test_instances_from_config = function()
    t.skip()

    local INSTANCE_1_FULLNAME = string.format('%s.instance-1', TEST_APP_NAME)
    local INSTANCE_2_FULLNAME = string.format('%s.instance-2', TEST_APP_NAME)
    local STATEBOARD_FULLNAME = string.format('%s-stateboard', TEST_APP_NAME)

    local CFG_PATH = fio.pathjoin(TEST_APP_DIR, 'instances.yml')

    local instances_cfg = table.concat({
       string.format('%s:', INSTANCE_1_FULLNAME),
       string.format('%s:', INSTANCE_2_FULLNAME)
    }, '\n')
    helper.write_file(CFG_PATH, instances_cfg)

    -- instance-1 - not started
    -- instance-2 - not started
    -- stateboard - not started
    check_is_not_running(INSTANCE_1_FULLNAME)
    check_is_not_running(INSTANCE_2_FULLNAME)
    check_is_not_running(STATEBOARD_FULLNAME)

    local status_strings = get_status_strings(TEST_APP_NAME)
    t.assert_equals(#status_strings, 2)
    t.assert_str_contains(status_strings[1], INSTANCE_1_FULLNAME)
    t.assert_str_contains(status_strings[1], 'Not started')
    t.assert_str_contains(status_strings[2], INSTANCE_2_FULLNAME)
    t.assert_str_contains(status_strings[2], 'Not started')

    local status_strings = get_status_strings(TEST_APP_NAME, {'--stateboard'})
    t.assert_equals(#status_strings, 3)
    t.assert_str_contains(status_strings[1], STATEBOARD_FULLNAME)
    t.assert_str_contains(status_strings[1], 'Not started')
    t.assert_str_contains(status_strings[2], INSTANCE_1_FULLNAME)
    t.assert_str_contains(status_strings[2], 'Not started')
    t.assert_str_contains(status_strings[3], INSTANCE_2_FULLNAME)
    t.assert_str_contains(status_strings[3], 'Not started')

    -- start instance-1 and stateboard
    start(INSTANCE_1_FULLNAME, {'--stateboard'})
    check_is_running(INSTANCE_1_FULLNAME)
    check_is_running(STATEBOARD_FULLNAME)

    -- instance-1 - running
    -- instance-2 - not started
    -- stateboard - running
    local status_strings = get_status_strings(TEST_APP_NAME)
    t.assert_equals(#status_strings, 2)
    t.assert_str_contains(status_strings[1], INSTANCE_1_FULLNAME)
    t.assert_str_contains(status_strings[1], 'Running')
    t.assert_str_contains(status_strings[2], INSTANCE_2_FULLNAME)
    t.assert_str_contains(status_strings[2], 'Not started')

    local status_strings = get_status_strings(TEST_APP_NAME, {'--stateboard'})
    t.assert_equals(#status_strings, 3)
    t.assert_str_contains(status_strings[1], STATEBOARD_FULLNAME)
    t.assert_str_contains(status_strings[1], 'Running')
    t.assert_str_contains(status_strings[2], INSTANCE_1_FULLNAME)
    t.assert_str_contains(status_strings[2], 'Running')
    t.assert_str_contains(status_strings[3], INSTANCE_2_FULLNAME)
    t.assert_str_contains(status_strings[3], 'Not started')

    -- stop instance-1
    stop(INSTANCE_1_FULLNAME)
    check_is_not_running(INSTANCE_1_FULLNAME)

    -- instance-1 - stopped
    -- instance-2 - not started
    -- stateboard - running
    local status_strings = get_status_strings(TEST_APP_NAME)
    t.assert_equals(#status_strings, 2)
    t.assert_str_contains(status_strings[1], INSTANCE_1_FULLNAME)
    t.assert_str_contains(status_strings[1], 'Stopped')
    t.assert_str_contains(status_strings[2], INSTANCE_2_FULLNAME)
    t.assert_str_contains(status_strings[2], 'Not started')

    local status_strings = get_status_strings(TEST_APP_NAME, {'--stateboard'})
    t.assert_equals(#status_strings, 3)
    t.assert_str_contains(status_strings[1], STATEBOARD_FULLNAME)
    t.assert_str_contains(status_strings[1], 'Running')
    t.assert_str_contains(status_strings[2], INSTANCE_1_FULLNAME)
    t.assert_str_contains(status_strings[2], 'Stopped')
    t.assert_str_contains(status_strings[3], INSTANCE_2_FULLNAME)
    t.assert_str_contains(status_strings[3], 'Not started')
end
