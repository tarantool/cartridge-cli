jit.off()

local fio = require('fio')

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

return helper
