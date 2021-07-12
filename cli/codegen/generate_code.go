package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/dave/jennifer/jen"
)

type generateLuaCodeOpts struct {
	PackageName  string
	FileName     string
	PackagePath  string
	VariablesMap map[string]string
}

var luaCodeFiles = []generateLuaCodeOpts{
	{
		PackageName: "admin",
		FileName:    "cli/admin/lua_code_gen.go",
		VariablesMap: map[string]string{
			"adminListFuncBodyTmpl":  "cli/admin/lua/admin_list_func_body_template.lua",
			"evalFuncGetResBodyTmpl": "cli/admin/lua/eval_func_get_res_body_template.lua",
		},
	},
	{
		PackageName: "connect",
		FileName:    "cli/connect/lua_code_gen.go",
		VariablesMap: map[string]string{
			"evalFuncBody":           "cli/connect/lua/eval_func_body.lua",
			"getSuggestionsFuncBody": "cli/connect/lua/get_suggestions_func_body.lua",
			"getTitleFuncBody":       "cli/connect/lua/get_title_func_body.lua",
		},
	},
	{
		PackageName: "connector",
		FileName:    "cli/connector/lua_code_gen.go",
		VariablesMap: map[string]string{
			"callFuncTmpl": "cli/connector/lua/call_func_template.lua",
			"evalFuncTmpl": "cli/connector/lua/eval_func_template.lua",
		},
	},
	{
		PackageName: "repair",
		FileName:    "cli/repair/lua_code_gen.go",
		VariablesMap: map[string]string{
			"reloadClusterwideConfigFuncBody": "cli/repair/lua/reload_clusterwide_config_func_body.lua",
		},
	},
	{
		PackageName: "replicasets",
		FileName:    "cli/replicasets/lua_code_gen.go",
		VariablesMap: map[string]string{
			"bootstrapVshardBody":                  "cli/replicasets/lua/bootstrap_vshard_body.lua",
			"getClusterIsHealthyBody":              "cli/replicasets/lua/get_cluster_is_healthy_body.lua",
			"editInstanceBody":                     "cli/replicasets/lua/edit_instance_body.lua",
			"editReplicasetsBodyTemplate":          "cli/replicasets/lua/edit_replicasets_body_template.lua",
			"formatTopologyReplicasetFuncTemplate": "cli/replicasets/lua/format_topology_replicaset_func_template.lua",
			"getKnownRolesBody":                    "cli/replicasets/lua/get_known_roles_body.lua",
			"getKnownVshardGroupsBody":             "cli/replicasets/lua/get_known_vshard_groups_body.lua",
			"getTopologyReplicasetsBodyTemplate":   "cli/replicasets/lua/get_topology_replicasets_body_template.lua",
		},
	},
	{
		PackageName: "cluster",
		FileName:    "cli/cluster/lua_code_gen.go",
		VariablesMap: map[string]string{
			"getMembershipInstancesBody": "cli/cluster/lua/get_membership_instances_body.lua",
			"probeInstancesBody":         "cli/cluster/lua/probe_instances_body.lua",
		},
	},
	{
		PackageName: "failover",
		FileName:    "cli/failover/lua_code_gen.go",
		VariablesMap: map[string]string{
			"setupFailoverBody": "cli/failover/lua/setup_failover_body.lua",
		},
	},
}

/* generateFileModeFile generates a file with map like this:

var FileModes = map[string]int{
	"filename": filemode,
	...
}
*/
func generateFileModeFile(path string, filename string) error {
	f := jen.NewFile("static")
	f.Comment("This file is generated! DO NOT EDIT\n")

	fileModeMap, err := getFileModes(path)

	if err != nil {
		return err
	}

	f.Var().Id("FileModes").Op("=").Map(jen.String()).Int().Values(jen.DictFunc(func(d jen.Dict) {
		for key, element := range fileModeMap {
			d[jen.Lit(key)] = jen.Lit(element).Commentf("/* %#o */", element)
		}
	}))

	f.Save(filename)

	return nil
}

func getFileModes(root string) (map[string]int, error) {
	fileModeMap := make(map[string]int)

	err := filepath.Walk(root, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fileInfo.IsDir() {
			rel, err := filepath.Rel(root, filePath)

			if err != nil {
				return err
			}

			fileModeMap[rel] = int(fileInfo.Mode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return fileModeMap, nil
}

func generateLuaCodeVar() error {
	for _, opts := range luaCodeFiles {
		f := jen.NewFile(opts.PackageName)
		f.Comment("This file is generated! DO NOT EDIT\n")

		for key, val := range opts.VariablesMap {
			content, err := ioutil.ReadFile(val)
			if err != nil {
				return err
			}

			f.Var().Id(key).Op("=").Lit(string(content))
		}

		f.Save(opts.FileName)
	}

	return nil
}

func main() {
	err := generateFileModeFile(
		"cli/create/templates/cartridge",
		"cli/codegen/static/create_cartrdige_template_filemodes_gen.go",
	)

	if err != nil {
		log.Errorf("Error while generating file modes: %s", err)
	}

	if err := generateLuaCodeVar(); err != nil {
		log.Errorf("Error while generating lua code string variables: %s", err)
	}
}
