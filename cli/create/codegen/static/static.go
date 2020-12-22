// +build dev

package static

//go:generate go run -tags=dev ../assets_generate.go

import "net/http"

var Data http.FileSystem = http.Dir("app")
