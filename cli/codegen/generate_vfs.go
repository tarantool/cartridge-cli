// +build ignore

package main

import (
	"github.com/apex/log"
	"github.com/shurcooL/vfsgen"
	"github.com/tarantool/cartridge-cli/cli/codegen/static"
)

func main() {
	err := vfsgen.Generate(static.CartridgeTemplateFS, vfsgen.Options{
		PackageName:  "create",
		BuildTags:    "!dev",
		VariableName: "CartridgeTemplateFS",
		Filename:     "../../create/vfsdata_gen.go",
	})

	if err != nil {
		log.Errorf("Error while generating cartridge virtual file system: %s", err)
	}

	err = vfsgen.Generate(static.AdminLuaTemplateFS, vfsgen.Options{
		PackageName:  "admin",
		BuildTags:    "!dev",
		VariableName: "AdminLuaTemplateFS",
		Filename:     "../../admin/vfsdata_gen.go",
	})

	if err != nil {
		log.Errorf("Error while generating cartridge virtual file system: %s", err)
	}
}
