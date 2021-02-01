
local res, err = {{ .FuncName }}(...)
assert(err == nil, err)
return res
