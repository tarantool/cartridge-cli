// +build dev

package static

//go:generate go run -tags=dev ../generate_vfs.go -o templatexxx.go

import (
	"net/http"
	"path/filepath"
)

var CreateCartridgeTemplateFS http.FileSystem = http.Dir(filepath.Join("..", "..", "create", "templates", "cartridge"))
var AdminLuaCodeFS http.FileSystem = http.Dir(filepath.Join("..", "..", "admin", "lua"))
var ConnectLuaCodeFS http.FileSystem = http.Dir(filepath.Join("..", "..", "connect", "lua"))
var ConnectorLuaCodeFS http.FileSystem = http.Dir(filepath.Join("..", "..", "connector", "lua"))
var RepairLuaCodeFS http.FileSystem = http.Dir(filepath.Join("..", "..", "repair", "lua"))
var ReplicasetsLuaCodeFS http.FileSystem = http.Dir(filepath.Join("..", "..", "replicasets", "lua"))
