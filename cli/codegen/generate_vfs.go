// +build ignore

package main

import (
	"net/http"

	"github.com/apex/log"
	"github.com/shurcooL/vfsgen"
	"github.com/tarantool/cartridge-cli/cli/codegen/static"
)

type generatedTemplate struct {
	FileSystem   http.FileSystem
	PackageName  string
	VariableName string
	FileName     string
}

var templates = []generatedTemplate{
	generatedTemplate{
		FileSystem:   static.CartridgeTemplateFS,
		PackageName:  "create",
		VariableName: "CartridgeTemplateFS",
		FileName:     "../../create/create_vfsdata_gen.go",
	},
	generatedTemplate{
		FileSystem:   static.AdminLuaTemplateFS,
		PackageName:  "admin",
		VariableName: "AdminLuaTemplateFS",
		FileName:     "../../admin/admin_vfsdata_gen.go",
	},
	generatedTemplate{
		FileSystem:   static.ConnectLuaTemplateFS,
		PackageName:  "connect",
		VariableName: "ConnectLuaTemplateFS",
		FileName:     "../../connect/connect_vfsdata_gen.go",
	},
	generatedTemplate{
		FileSystem:   static.ConnectorLuaTemplateFS,
		PackageName:  "connector",
		VariableName: "ConnectorLuaTemplateFS",
		FileName:     "../../connector/connector_vfsdata_gen.go",
	},
	generatedTemplate{
		FileSystem:   static.ConnectLuaTemplateFS,
		PackageName:  "repair",
		VariableName: "RepairLuaTemplateFS",
		FileName:     "../../repair/repair_vfsdata_gen.go",
	},
	generatedTemplate{
		FileSystem:   static.ReplicasetsLuaTemplateFS,
		PackageName:  "replicasets",
		VariableName: "ReplicasetsLuaTemplateFS",
		FileName:     "../../replicasets/replicasets_vfsdata_gen.go",
	},
}

func main() {
	for _, tmpl := range templates {
		err := vfsgen.Generate(tmpl.FileSystem, vfsgen.Options{
			PackageName:  tmpl.PackageName,
			BuildTags:    "!dev",
			VariableName: tmpl.VariableName,
			Filename:     tmpl.FileName,
		})

		if err != nil {
			log.Errorf("Error while generating file system: %s", err)
		}
	}
}
