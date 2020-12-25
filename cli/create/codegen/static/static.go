// +build dev

package static

//go:generate go run -tags=dev ../generate_vfs.go

import "net/http"

var CartridgeData http.FileSystem = http.Dir("../../templates/cartridge")
