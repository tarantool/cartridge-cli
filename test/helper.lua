jit.off()

local fio = require('fio')
local clock = require('clock')
local fiber = require('fiber')
local ffi = require('ffi')
local fun = require('fun')

local Process = require('luatest').Process

-- box.NULL, custom and cdata errors aware assert
function assert(val, message, ...) -- luacheck: no global
    if not val or val == nil then
        error(tostring(message), 2)
    end
    return val, message, ...
end

local helper = {}

helper.tempdir = 'tmp'

local function build_cli_binary()
    local cli_src_path = fio.abspath('.')
    local cli_build_dir = fio.pathjoin(helper.tempdir, 'cli-build')

    if fio.path.exists(cli_build_dir) then
        assert(fio.rmtree(cli_build_dir))
    end

    assert(fio.mktree(cli_build_dir))

    local cmd = string.format(
        'cd %s && tarantoolctl rocks make --chdir %s',
        cli_build_dir,
        cli_src_path
    )
    local rc = os.execute(cmd)
    assert(rc == 0)

    local cli_binary_path = fio.pathjoin(cli_build_dir, '.rocks/bin/cartridge')
    assert(fio.path.exists(cli_binary_path))

    return cli_binary_path
end

helper.cartridge_cmd = build_cli_binary()

function helper.check_pid_running(pid)
    return ffi.C.kill(tonumber(pid), 0) == 0
end

function helper.wait_process_exit(pid, timeout)
    timeout = timeout or 2
    if type(pid) == 'table' then
        pid = tonumber(pid.pid)
    end
    local started_at = clock.time()
    while helper.check_pid_running(pid) do
        if clock.time() - started_at > timeout then
            error('expected process to exit, but it does not')
        end
        fiber.sleep(0.1)
    end
end

-- Non-blocking os.execute() which fails if process does not exit.
function helper.os_execute(path, args, opts)
    opts = opts or {}
    local env = fun.chain(os.environ(), opts.env or {}):tomap()

    local process = Process:start(fio.abspath(path), args, env)
    helper.wait_process_exit(process, opts.timeout)
    return process
end

function helper.read_file(path)
    local file = assert(fio.open(path))
    local result = assert(file:read())
    file:close()
    return result
end

function helper.concat(...)
    return fun.chain(...):totable()
end

return helper
