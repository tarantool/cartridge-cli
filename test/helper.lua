package.path = package.path .. ";./?;?"

-- box.NULL, custom and cdata errors aware assert
function assert(val, message, ...) -- luacheck: no global
    if not val or val == nil then
        error(tostring(message), 2)
    end
    return val, message, ...
end
