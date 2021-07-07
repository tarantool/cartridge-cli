package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	lua "github.com/yuin/gopher-lua"
)

const (
	rocksManifestPath = ".rocks/share/tarantool/rocks/manifest"
)

// LuaReadStringVar reads global string variable from specified Lua file
func LuaReadStringVar(filePath string, varName string) (string, error) {
	L := lua.NewState()
	defer L.Close()

	// set env to empty table
	emptyEnv := lua.LTable{}
	L.Env = &emptyEnv

	if err := L.DoFile(filePath); err != nil {
		return "", fmt.Errorf("Failed to read file %s: %s", filePath, err)
	}

	luaVal := L.Env.RawGetString(varName)
	if luaVal.Type() == lua.LTNil {
		return "", fmt.Errorf("Variable `%s` is not set in %s", varName, filePath)
	}

	if luaVal.Type() != lua.LTString {
		return "", fmt.Errorf("Field `%s` must be string in %s", varName, filePath)
	}

	return luaVal.String(), nil
}

// LuaGetRocksVersionsAndDuplicates gets {name: version} map from rocks manifest
func LuaGetRocksVersionsAndDuplicates(appDirPath string) (map[string]string, []string, error) {
	rocksVersionsMap := map[string]string{}
	rocksDuplicates := []string{}

	manifestFilePath := filepath.Join(appDirPath, rocksManifestPath)
	if _, err := os.Stat(manifestFilePath); err == nil {
		L := lua.NewState()
		defer L.Close()

		if err := L.DoFile(manifestFilePath); err != nil {
			return nil, nil, fmt.Errorf("Failed to read manifest file %s: %s", manifestFilePath, err)
		}

		depsL := L.Env.RawGetString("dependencies")
		depsLTable, ok := depsL.(*lua.LTable)
		if !ok {
			return nil, nil, fmt.Errorf("Failed to read manifest file: dependencies is not a table")
		}

		depsLTable.ForEach(func(depNameL lua.LValue, depInfoL lua.LValue) {
			depName := depNameL.String()

			depInfoLTable, ok := depInfoL.(*lua.LTable)
			if !ok {
				log.Warnf("Failed to get %s dependency info", depName)
			} else {
				depInfoLTable.ForEach(func(depVersionL lua.LValue, _ lua.LValue) {
					depVersion := depVersionL.String()
					if _, found := rocksVersionsMap[depName]; found {
						rocksDuplicates = append(rocksDuplicates, depName)
					}
					rocksVersionsMap[depName] = depVersion
				})
			}
		})

	} else if !os.IsNotExist(err) {
		return nil, nil, fmt.Errorf("Failed to read manifest file %s: %s", manifestFilePath, err)
	}

	return rocksVersionsMap, rocksDuplicates, nil
}
