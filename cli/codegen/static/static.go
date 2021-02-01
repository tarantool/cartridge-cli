// +build dev

package static

//go:generate go run -tags=dev ../generate_vfs.go -o templatexxx.go

import (
	"net/http"
	"path/filepath"
)

var CartridgeTemplateFS http.FileSystem = http.Dir(filepath.Join("..", "..", "create", "templates", "cartridge"))
var AdminLuaTemplateFS http.FileSystem = http.Dir(filepath.Join("..", "..", "admin", "templates"))
var ConnectLuaTemplateFS http.FileSystem = http.Dir(filepath.Join("..", "..", "connect", "templates"))
var ConnectorLuaTemplateFS http.FileSystem = http.Dir(filepath.Join("..", "..", "connector", "templates"))
var RepairLuaTemplateFS http.FileSystem = http.Dir(filepath.Join("..", "..", "repair", "templates"))
var ReplicasetsLuaTemplateFS http.FileSystem = http.Dir(filepath.Join("..", "..", "replicasets", "templates"))
