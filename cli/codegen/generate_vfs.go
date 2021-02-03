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

var templates = []generateFSOpts{
	generateFSOpts{
		FileSystem:   static.CreateCartridgeTemplateFS,
		PackageName:  "create",
		VariableName: "CreateCartridgeTemplateFS",
		FileName:     "../../create/create_vfsdata_gen.go",
	},
	/*
		generateFSOpts{
			FileSystem:   static.AdminLuaCodeFS,
			PackageName:  "admin",
			VariableName: "AdminLuaCodeFS",
			FileName:     "../../admin/admin_vfsdata_gen.go",
		},
		generateFSOpts{
			FileSystem:   static.ConnectLuaCodeFS,
			PackageName:  "connect",
			VariableName: "ConnectLuaCodeFS",
			FileName:     "../../connect/connect_vfsdata_gen.go",
		},
		generateFSOpts{
			FileSystem:   static.ConnectorLuaCodeFS,
			PackageName:  "connector",
			VariableName: "ConnectorLuaCodeFS",
			FileName:     "../../connector/connector_vfsdata_gen.go",
		},
		generateFSOpts{
			FileSystem:   static.RepairLuaCodeFS,
			PackageName:  "repair",
			VariableName: "RepairLuaCodeFS",
			FileName:     "../../repair/repair_vfsdata_gen.go",
		},
		generateFSOpts{
			FileSystem:   static.ReplicasetsLuaCodeFS,
			PackageName:  "replicasets",
			VariableName: "ReplicasetsLuaCodeFS",
			FileName:     "../../replicasets/replicasets_vfsdata_gen.go",
		},
	*/
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
