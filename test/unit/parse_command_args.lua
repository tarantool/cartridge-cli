local t = require('luatest')
local app = require('cartridge-cli')

local g = t.group('parse_command_args')

local function split(msg)
    return msg:split(' ')
end

g.test_parse_empty = function()
    t.assert_equals(app.parse({}, {}), {})
end

g.test_simple_schema = function()
    local schema = {
        opts = {
            name = 'string',
            count = 'number',
        },
        args = {
            'type',
            'path',
        }
    }

    t.assert_equals(app.parse({}, schema), {})
    t.assert_equals(app.parse(split('--name NAME'), schema), { name = 'NAME' })
    t.assert_equals(app.parse(split('TYPE'), schema), { type = 'TYPE' })
    t.assert_equals(app.parse(split('--name NAME TYPE'), schema), { type = 'TYPE', name = 'NAME' })
    t.assert_equals(
        app.parse(split('--name NAME --count 1 TYPE'), schema),
        { type = 'TYPE', name = 'NAME', count = 1 }
    )
    t.assert_equals(
        app.parse(split('--name=NAME --count=1 TYPE'), schema),
        { type = 'TYPE', name = 'NAME', count = 1 }
    )
    t.assert_equals(
        app.parse(split('--name=NAME --count=1 TYPE PATH'), schema),
        { path = 'PATH', name = 'NAME', count = 1, type = 'TYPE' }
    )
    t.assert_equals(
        app.parse(split('--name=NAME --count=1 TYPE PATH'), schema),
        { path = 'PATH', name = 'NAME', count = 1, type = 'TYPE' }
    )
    t.assert_equals(
        app.parse(split('--name=NAME TYPE --count 1 PATH'), schema),
        { path = 'PATH', name = 'NAME', count = 1, type = 'TYPE' }
    )
    t.assert_equals(
        app.parse(split('TYPE PATH --name NAME --count 1'), schema),
        { path = 'PATH', name = 'NAME', count = 1, type = 'TYPE' }
    )

    local res, err = app.parse(split('TYPE PATH UNKNOWN'), schema)
    t.assert_equals(res, nil)
    t.assert_str_icontains(err, 'unknown option')

    local res, err = app.parse(split('--unknown UNKNOWN'), schema)
    t.assert_equals(res, nil)
    t.assert_str_icontains(err, 'unknown option')

    local res, err = app.parse(split('--name NAME1 --name NAME2'), schema)
    t.assert_equals(res, nil)
    t.assert_str_icontains(err, 'option name passed more than one time')

    local res, err = app.parse({}, { args = { arg = 'string' } })
    t.assert_equals(res, nil)
    t.assert_str_icontains (err, 'args should be an array')
end


g.test_schema_with_flag = function()
    local schema = {
        opts = {
            name = 'string',
            flag = 'boolean',
        },
        args = {
            'type',
            'path',
        }
    }

    t.assert_equals(app.parse(split('--flag'), schema), { flag = true })
    t.assert_equals(app.parse(split('--flag=1'), schema), { flag = true })
    t.assert_equals(app.parse(split('--flag=true'), schema), { flag = true })
    t.assert_equals(app.parse(split('--flag 0'), schema), { flag = false })
    t.assert_equals(app.parse(split('--flag false'), schema), { flag = false })

    t.assert_equals(app.parse(split('--name NAME --flag'), schema), { name = 'NAME', flag = true })
    t.assert_equals(
        app.parse(split('--name NAME --flag TYPE'), schema),
        { type = 'TYPE', name = 'NAME', flag = true }
    )
    t.assert_equals(
        app.parse(split('--name NAME TYPE --flag'), schema),
        { type = 'TYPE', name = 'NAME', flag = true }
    )
    t.assert_equals(
        app.parse(split('--flag --name NAME TYPE'), schema),
        { type = 'TYPE', name = 'NAME', flag = true }
    )

    t.assert_equals(
        app.parse(split('--name NAME --flag false TYPE'), schema),
        { type = 'TYPE', name = 'NAME', flag = false }
    )
    t.assert_equals(
        app.parse(split('--name NAME TYPE --flag false'), schema),
        { type = 'TYPE', name = 'NAME', flag = false }
    )
    t.assert_equals(
        app.parse(split('--flag false --name NAME TYPE'), schema),
        { type = 'TYPE', name = 'NAME', flag = false }
    )

    t.assert_equals(
        app.parse(split('--flag TYPE PATH'), schema),
        { type = 'TYPE', flag = true, path = 'PATH' }
    )
    t.assert_equals(
        app.parse(split('TYPE --flag PATH'), schema),
        { type = 'TYPE', flag = true, path = 'PATH' }
    )
    t.assert_equals(
        app.parse(split('TYPE PATH --flag'), schema),
        { type = 'TYPE', flag = true, path = 'PATH' }
    )

    t.assert_equals(
        app.parse(split('--flag false TYPE PATH'), schema),
        { type = 'TYPE', flag = false, path = 'PATH' }
    )
    t.assert_equals(
        app.parse(split('TYPE --flag false PATH'), schema),
        { type = 'TYPE', flag = false, path = 'PATH' }
    )
    t.assert_equals(
        app.parse(split('TYPE PATH --flag false'), schema),
        { type = 'TYPE', flag = false, path = 'PATH' }
    )

    t.assert_equals(
        app.parse(split('true false --flag false'), schema),
        { type = 'true', flag = false, path = 'false' }
    )
    t.assert_equals(
        app.parse(split('true --flag false false'), schema),
        { type = 'true', flag = false, path = 'false' }
    )
    t.assert_equals(
        app.parse(split('--flag false true false'), schema),
        { type = 'true', flag = false, path = 'false' }
    )
end


g.test_schema_with_two_flags = function()
    local schema = {
        opts = {
            name = 'string',
            flag1 = 'boolean',
            flag2 = 'boolean',
        },
        args = {
            'type',
            'path',
        }
    }

    t.assert_equals(app.parse(split('--flag1'), schema), { flag1 = true })
    t.assert_equals(app.parse(split('--flag1 --flag2'), schema), { flag1 = true, flag2 = true })
    t.assert_equals(app.parse(split('--flag1 false --flag2'), schema), { flag1 = false, flag2 = true })
    t.assert_equals(app.parse(split('--flag1 --flag2 false'), schema), { flag1 = true, flag2 = false })

    t.assert_equals(
        app.parse(split('TYPE --name NAME --flag1 --flag2'), schema),
        { type = 'TYPE', name = 'NAME', flag1 = true, flag2 = true }
    )
    t.assert_equals(
        app.parse(split('TYPE --flag1 --name NAME --flag2'), schema),
        { type = 'TYPE', name = 'NAME', flag1 = true, flag2 = true }
    )
    t.assert_equals(
        app.parse(split('TYPE --flag1 --flag2 --name NAME'), schema),
        { type = 'TYPE', name = 'NAME', flag1 = true, flag2 = true }
    )
    t.assert_equals(
        app.parse(split('--flag1 TYPE --flag2 --name NAME'), schema),
        { type = 'TYPE', name = 'NAME', flag1 = true, flag2 = true }
    )
    t.assert_equals(
        app.parse(split('--flag1 --flag2 TYPE --name NAME'), schema),
        { type = 'TYPE', name = 'NAME', flag1 = true, flag2 = true }
    )

    local res, err = app.parse(split('--flag1 --flag1'), schema)
    t.assert_equals(res, nil)
    t.assert_str_icontains(err, 'option flag1 passed more than one time')
end

g.test_prettifyed_opts = function()
    local res, err = app.parse({}, { opts = { ['long-option'] = 'string' } })
    t.assert_equals(res, nil)
    t.assert_str_icontains (err, 'option name can not contain "-" symbol')

    local schema = {
        opts = {
            long_option = 'string',
            long_flag = 'boolean',
        },
        args = {
            'path',
        }
    }

    t.assert_equals(
        app.parse(split('--long-option VALUE'), schema),
        { long_option = 'VALUE' }
    )
    t.assert_equals(
        app.parse(split('--long_option VALUE'), schema),
        { long_option = 'VALUE' }
    )
    t.assert_equals(
        app.parse(split('--long_flag'), schema),
        { long_flag = true }
    )
    t.assert_equals(
        app.parse(split('--long-flag'), schema),
        { long_flag = true }
    )
    t.assert_equals(
        app.parse(split('--long-flag PATH'), schema),
        { long_flag = true, path = 'PATH' }
    )

    t.assert_equals(
        app.parse(split('--long-flag --long-option VALUE PATH'), schema),
        { long_flag = true, long_option = 'VALUE', path = 'PATH' }
    )
    t.assert_equals(
        app.parse(split('--long-option VALUE --long-flag PATH'), schema),
        { long_flag = true, long_option = 'VALUE', path = 'PATH' }
    )
    t.assert_equals(
        app.parse(split('--long-option VALUE PATH --long-flag'), schema),
        { long_flag = true, long_option = 'VALUE', path = 'PATH' }
    )

    local res, err = app.parse(
        split('--long_option VALUE --long-option VALUE'),
        schema
    )
    t.assert_equals(res, nil)
    t.assert_str_icontains (err, 'option long_option passed more than one time')
end
