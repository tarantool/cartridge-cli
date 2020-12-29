// +build dev

package static

//go:generate go run -tags=dev ../generate_vfs.go -o templatexxx.go

import (
	"net/http"
	"path/filepath"
)

var CartridgeTemplateFS http.FileSystem = http.Dir(filepath.Join("..", "..", "templates", "cartridge"))
