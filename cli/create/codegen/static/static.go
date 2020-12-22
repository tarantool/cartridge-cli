// +build dev

package static

//go:generate go run -tags=dev ../generate_code.go

import "net/http"

var Data http.FileSystem = http.Dir("content")
