// +build ignore

package main

import (
	"net/http"

	"github.com/apex/log"
	"github.com/shurcooL/vfsgen"
	"github.com/tarantool/cartridge-cli/cli/codegen/static"
)

type generateFSOpts struct {
	FileSystem   http.FileSystem
	PackageName  string
	VariableName string
	FileName     string
}

var FSOpts = []generateFSOpts{
	generateFSOpts{
		FileSystem:   static.CreateCartridgeTemplateFS,
		PackageName:  "create",
		VariableName: "CreateCartridgeTemplateFS",
		FileName:     "../../create/create_vfsdata_gen.go",
	},
}

func main() {
	for _, opts := range FSOpts {
		err := vfsgen.Generate(opts.FileSystem, vfsgen.Options{
			PackageName:  opts.PackageName,
			BuildTags:    "!dev",
			VariableName: opts.VariableName,
			Filename:     opts.FileName,
		})

		if err != nil {
			log.Errorf("Error while generating file system: %s", err)
		}
	}
}
