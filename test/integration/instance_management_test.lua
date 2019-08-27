local t = require('luatest')
local g = t.group('instance_managemnet')

local clock = require('clock')
local fiber = require('fiber')
local fio = require('fio')
local fun = require('fun')

local cmd = assert(package.search('cartridge'))

local function check_pid_running(pid)
    return os.execute('ps -p ' .. pid .. ' > /dev/null') == 0
end

local function wait_process_exit(pid, timeout)
    timeout = timeout or 2
    if type(pid) == 'table' then
        pid = tonumber(pid.pid)
    end
    local started_at = clock.time()
    while check_pid_running(pid) do
        if clock.time() - started_at > timeout then
            error('expected process to exit, but it does not')
        end
        fiber.sleep(0.1)
    end
end

-- Non-blocking os.execute() which fails if process does not exit.
local function os_execute(cmd, args, env)
    env = fun.chain(os.environ(), env or {}):tomap()
    local process = t.Process:start(fio.abspath(cmd), args, env)
    wait_process_exit(process)
    return process
end

local function read_file(path)
    local file = assert(fio.open(path))
    local result = assert(file:read())
    file:close()
    return result
end

local function concat(...)
    return fun.chain(...):totable()
end

local RUN_DIR = 'tmp/test_run'
local TEST_OPTS = {'--run_dir', RUN_DIR}
local SIMPLE_INSTANCE_OPTS = concat({'test_name', '--script', 'test/instances/simple.lua'}, TEST_OPTS)

g.setup = function() fio.rmtree(RUN_DIR) end

g.test_start_stop = function()
    local starter = os_execute(cmd, concat({'start'}, SIMPLE_INSTANCE_OPTS))
    require('log').info({o = starter})
    local pid = read_file('tmp/test_run/test_name.pid')
    t.assert_not_equals(pid, starter.pid)
    t.assert(check_pid_running(pid))
    os_execute(cmd, concat({'stop', 'test_name'}, TEST_OPTS))
    t.assert_not(check_pid_running(pid))
    t.assert_not(fio.stat('tmp/test_run/test_name.pid'))
end

g.test_start_foreground = function()
    local starter = t.Process:start(
        cmd,
        concat({'start'}, SIMPLE_INSTANCE_OPTS, {'--foreground'}),
        os.environ()
    )
    local pid = t.helpers.retrying({}, function()
        return tonumber(read_file('tmp/test_run/test_name.pid'))
    end)
    t.assert_equals(pid, starter.pid)
    t.assert(check_pid_running(pid))
    starter:kill()
    t.helpers.retrying({}, function() t.assert_not(check_pid_running(pid)) end)
end
