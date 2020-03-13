local fio = require('fio')
local t = require('luatest')
local app = require('cartridge-cli')

local g = t.group('sdk_path_detection')

local detect_sdk_path = app.detect_sdk_path

g.test_empty_args = function()
    local res, err = detect_sdk_path({})
    t.assert_equals(res, nil)
    t.assert_str_icontains(err, 'you should specify one of')
    t.assert_str_icontains(err, '--sdk-local')
    t.assert_str_icontains(err, '--sdk-path')
end

g.test_both_args = function()
    local SDK_PATH = 'SDK_PATH'

    local res, err = detect_sdk_path({sdk_local = true, sdk_path = SDK_PATH})

    t.assert_equals(res, nil)
    t.assert_str_icontains(err, 'you should specify one of')
    t.assert_str_icontains(err, '--sdk-local')
    t.assert_str_icontains(err, '--sdk-path')
end

g.test_sdk_path_arg_passed = function()
    local SDK_PATH = 'SDK_PATH'

    local res, err = detect_sdk_path({sdk_path = SDK_PATH})
    t.assert_equals(err, nil)
    t.assert_equals(res, SDK_PATH)

    local SDK_PATH = 'SDK_PATH'

    local res, err = detect_sdk_path({sdk_local = false, sdk_path = SDK_PATH})
    t.assert_equals(err, nil)
    t.assert_equals(res, SDK_PATH)
end

g.test_sdk_local_arg_passed = function()
    local TARANTOOL_PATH = fio.abspath(fio.dirname(arg[-1]))

    local res, err = detect_sdk_path({sdk_local = true})
    t.assert_equals(err, nil)
    t.assert_equals(res, TARANTOOL_PATH)
end
