// +build ignore

package main

import (
	"github.com/apex/log"
	"github.com/shurcooL/vfsgen"
	"github.com/tarantool/cartridge-cli/cli/create/codegen/static"
)

func main() {
	err := vfsgen.Generate(static.CartridgeTemplateFS, vfsgen.Options{
		PackageName:  "static",
		BuildTags:    "!dev",
		VariableName: "CartridgeTemplateFS",
		Filename:     "cartridge_vfsdata_gen.go",
	})

	if err != nil {
		log.Errorf("Error while generating static files assets: %s", err)
	}
}
