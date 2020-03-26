local argparse_internal = require('internal.argparse')

local utils = require('cartridge-cli.utils')

local argparse = {}

local BOOLEAN_VALUES = {'0', '1', 'true', 'false'}

local function is_option_name(arg)
    return string.startswith(arg, '-')
end

local function is_option_vith_value(arg)
    -- --opt=value
    return string.startswith(arg, '-') and string.find(arg, '=') ~= nil
end

local function raw_option_name(arg)
    assert(is_option_name(arg))
    return arg:gsub('^%-%-?', ''):gsub('-', '_')
end

local function prettify_option_name(opt_name)
    local pretyy_opt_name = opt_name:gsub('_', '-')
    return pretyy_opt_name
end

function argparse.parse(args, schema)
    --[[
        <command> --name NAME --debug TYPE PATH

        schema = {
            opts = {
                name = 'string',
                debug = 'boolean',
            },
            args = {
                'type',
                'path',
            }
        }

        Both `--long-opt OPT` and `--long_opt OPT` will be parsed as `{ long_opt = 'OPT' }`
        You can't define option `long-opt` in schema, only `long_opt` pattern is allowed.
    --]]

    args = args or {}
    schema = schema or {}

    local schema_opts = schema.opts or {}
    local schema_args = schema.args or {}

    -- Validate schema
    -- - check that schema.args is an array, not a map
    -- - and schema doesn't contain args and options with the same names
    for key, arg in pairs(schema_args) do
        if type(key) ~= 'number' then
            return nil, string.format('schema.args should be an array, not a map')
        end
        if schema_opts[arg] ~= nil then
            return nil, string.format('Defined arg and option with the same name: %s', arg)
        end
    end

    -- - check that schema options use "long_opt" pattern, not "long-opt"
    for opt_name, _ in pairs(schema_opts) do
        if opt_name:find('-') ~= nil then
            local err = string.format(
                'Option name can not contain "-" symbol (got %s). ' ..
                    'Please, use "long_opt" pattern instead of "long-opt" in arguments schema',
                opt_name
            )
            return nil, err
        end
    end

    -- Split options
    -- - replace {'--opt=value'} with {'--opt', 'value'}
    local splitted_args = {}
    for _, arg in ipairs(args) do
        if is_option_vith_value(arg) then
            local opt, value = unpack(arg:split('=', 1))

            table.insert(splitted_args, opt)
            table.insert(splitted_args, value)
        else
            table.insert(splitted_args, arg)
        end
    end

    args = splitted_args

    -- Validate args
    -- - check that all options are mentioned no more than one time
    local passed_opts = {}
    for _, arg in ipairs(args) do
        if is_option_name(arg) then
            local option_name = raw_option_name(arg)

            if passed_opts[option_name] then
                return nil, string.format('Option %s passed more than one time', option_name)
            end

            passed_opts[option_name] = true
        end
    end

    -- Collect boolean options
    -- - since argparse works different for different Tarantool versions
    -- - (it concerns using --flag=<value> for boolean options)
    -- - boolean options are parsed separately
    local indexes_to_keep = {}
    local boolean_ops = {}

    for i, arg in ipairs(args) do
        if not is_option_name(arg) then
            if indexes_to_keep[i] == nil then
                indexes_to_keep[i] = true
            end
        else
            local option_name = raw_option_name(arg)

            if schema_opts[option_name] ~= 'boolean' then
                indexes_to_keep[i] = true
            else
                indexes_to_keep[i] = false
                if not utils.array_contains(BOOLEAN_VALUES, args[i + 1]) then
                    -- if next value isn't boolean then flag is set
                    boolean_ops[option_name] = true
                else
                    -- if next value is boolean value, skip it
                    indexes_to_keep[i + 1] = false
                    if args[i + 1] == 'true' or tostring(args[i + 1]) == '1' then
                        -- flag is set
                        boolean_ops[option_name] = true
                    else
                        -- flag isn't set
                        boolean_ops[option_name] = false
                    end
                end
            end
        end
    end

    local normalized_args = {}
    for i, arg in ipairs(args) do
        if indexes_to_keep[i] then
            table.insert(normalized_args, arg)
        end
    end

    args = normalized_args

    -- Prepare args for `internal.argparse`
    -- - convert options to `internal.argparse` format
    -- - it should be able to parse --long-name as well as --long_name
    local argparse_opts = {}

    for opt_name, opt_type in pairs(schema_opts) do
        table.insert(argparse_opts, {opt_name, opt_type})
        if opt_name:find('_') ~= nil then
            local pretty_opt_name = prettify_option_name(opt_name)
            table.insert(argparse_opts, {pretty_opt_name, opt_type})
        end
    end

    -- Call `internal.argparse.parse()`
    local ok, parsed_parameters = pcall(function()
        return argparse_internal.parse(args, argparse_opts)
    end)
    if not ok then
        return nil, string.format('Parse error: %s', parsed_parameters)
    end

    -- Construct result
    local res = {}

    -- - collect args
    for i, arg_name in pairs(schema_args) do
        if parsed_parameters[i] ~= nil then
            res[arg_name] = parsed_parameters[i]
            parsed_parameters[i] = nil
        end
    end

    -- - collect opts
    for parsed_opt_name, opt_value in pairs(parsed_parameters) do
        local opt_name = string.gsub(parsed_opt_name, '-', '_')
        if schema_opts[opt_name] ~= nil then
            res[opt_name] = opt_value
        else
            return nil, string.format('Unknown option: %s', opt_name)
        end
    end

    -- - collect boolean opts
    for parsed_opt_name, opt_value in pairs(boolean_ops) do
        local opt_name = string.gsub(parsed_opt_name, '-', '_')
        if schema_opts[opt_name] ~= nil then
            res[opt_name] = opt_value
        else
            return nil, string.format('Unknown option: %s', opt_name)
        end
    end

    return res
end

return argparse
