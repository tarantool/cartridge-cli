
local func_list, err = {{ .AdminListFuncName }}(...)
assert(err == nil, err)
return func_help
