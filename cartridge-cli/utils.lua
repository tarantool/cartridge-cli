local fio = require('fio')
local errno = require('errno')

local utils = {}

function utils.remove_leading_spaces(s, spaces_num)
    spaces_num = spaces_num or 8
    local REMOVE_PATTERN = string.format('^%s', string.rep(' ', spaces_num))

    local res_lines = {}
    for _, line in ipairs(s:split('\n')) do
        local res_line = line:gsub(REMOVE_PATTERN, '')
        table.insert(res_lines, res_line)
    end

    return table.concat(res_lines, '\n'):strip()
end

function utils.make_tree(path)
    local ok, err = fio.mktree(path)
    if not ok then
        return false, string.format("Failed to create path %s: %s", path, err)
    end
    return true
end

function utils.write_file(path, data, mode)
    mode = mode or tonumber(644, 8)

    local file = fio.open(path, {'O_CREAT', 'O_WRONLY', 'O_TRUNC', 'O_SYNC'}, mode)
    if file == nil then
        return false, string.format('Failed to open file %s: %s', path, errno.strerror())
    end

    local res = file:write(data)

    if not res then
        return false, string.format('Failed to write to file %s: %s', path, errno.strerror())
    end

    file:close()

    return true
end

function utils.merge_lists(...)
    local res = {}
    for i = 1, select('#', ...) do
        local t = select(i, ...)
        for _, v in ipairs(t) do
            res[#res + 1] = v
        end
    end
    return res
end

-- expand() allows to render a text template, expanding ${statement}
-- into the calculated value of that statement.
-- Roughly based on http://lua-users.org/wiki/TextTemplate
--
-- First argument is the template string, then arbitrary number of
-- tables with mappings of variable=value
function utils.expand(template, ...)
    assert(type(template)=='string', 'expecting string')
    local searchlist = {...}
    local estring,evar

    local statements = {'do', 'if', 'for', 'while', 'repeat'}

    function estring(str)
        local b,e,i
        b,i = string.find(str, '%$.')
        if not b then return str end

        local R, pos = {}, 1
        repeat
            b,e = string.find(str, '^%b{}', i)
            if b then
                table.insert(R, string.sub(str, pos, b-2))
                table.insert(R, evar(string.sub(str, b+1, e-1)))
                i = e+1
                pos = i
            else
                b,e = string.find(str, '^%b()', i)
                if b then
                    table.insert(R, string.sub(str, pos, b-2))
                    table.insert(R, evar(string.sub(str, b+1, e-1)))
                    i = e+1
                    pos = i
                elseif string.find(str, '^%a', i) then
                    table.insert(R, string.sub(str, pos, i-2))
                    table.insert(R, evar(string.sub(str, i, i)))
                    i = i+1
                    pos = i
                elseif string.find(str, '^%$', i) then
                    table.insert(R, string.sub(str, pos, i))
                    i = i+1
                    pos = i
                end
            end
            b,i = string.find(str, '%$.', i)
        until not b

        table.insert(R, string.sub(str, pos))
        return table.concat(R)
    end

    local function search(index)
        for _,symt in ipairs(searchlist) do
            local ts = type(symt)
            local value
            if     ts == 'function' then value = symt(index)
            elseif ts == 'table'
            or ts == 'userdata' then value = symt[index]
                if type(value)=='function' then value = value(symt) end
            else error'search item must be a function, table or userdata' end
            if value ~= nil then return value end
        end
        error('unknown variable: '.. index)
    end

    local function elist(var, v, str, sep)
        local tab = search(v)
        if tab then
            assert(type(tab)=='table', 'expecting table from: '.. var)
            local R = {}
            table.insert(searchlist, 1, tab)
            table.insert(searchlist, 1, false)
            for _,elem in ipairs(tab) do
                searchlist[1] = elem
                table.insert(R, estring(str))
            end
            table.remove(searchlist, 1)
            table.remove(searchlist, 1)
            return table.concat(R, sep)
        else
            return ''
        end
    end

    local function get(tab,index)
        for _,symt in ipairs(searchlist) do
            local ts = type(symt)
            local value
            if     ts == 'function' then value = symt(index)
            elseif ts == 'table'
            or ts == 'userdata' then value = symt[index]
            else error'search item must be a function, table or userdata' end
            if value ~= nil then
                tab[index] = value  -- caches value and prevents changing elements
                return value
            end
        end
    end

    function evar(var)
        if string.find(var, '^[_%a][_%w]*$') then -- ${vn}
            return estring(tostring(search(var)))
        end
        local _,e,cmd = string.find(var, '^(%a+)%s.')
        if cmd == 'foreach' then -- ${foreach vn xxx} or ${foreach vn/sep/xxx}
            local vn,s
            _,e,vn,s = string.find(var, '^([_%a][_%w]*)([%s%p]).', e)
            if vn then
                if string.find(s, '%s') then
                    return elist(var, vn, string.sub(var, e), '')
                end
                local b = string.find(var, s, e, true)
                if b then
                    return elist(var, vn, string.sub(var, b+1), string.sub(var,e,b-1))
                end
            end
            error('syntax error in: '.. var, 2)
        elseif cmd == 'when' then -- $(when vn xxx)
            local vn
            _,e,vn = string.find(var, '^([_%a][_%w]*)%s.', e)
            if vn then
                local t = search(vn)
                if not t then
                    return ''
                end
                local s = string.sub(var,e)
                if type(t)=='table' or type(t)=='userdata' then
                    table.insert(searchlist, 1, t)
                    s = estring(s)
                    table.remove(searchlist, 1)
                    return s
                else
                    return estring(s)
                end
            end
            error('syntax error in: '.. var, 2)
        else
            if statements[cmd] then -- do if for while repeat
                var = 'local OUT="" '.. var ..' return OUT'
            else  -- expression
                var = 'return '.. var
            end
            local f = assert(loadstring(var))
            local t = searchlist[1]
            assert(type(t)=='table' or type(t)=='userdata', 'expecting table')
            setfenv(f, setmetatable({}, {__index=get, __newindex=t}))
            return estring(tostring(f()))
        end
    end

    return estring(template)
end


return utils
