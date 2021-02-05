// +build dev

package static

//go:generate go run -tags=dev ../generate_templates_vfs.go -o templatexxx.go

import (
	"net/http"
	"path/filepath"
)

var CreateCartridgeTemplateFS http.FileSystem = http.Dir(filepath.Join("..", "..", "create", "templates", "cartridge"))
