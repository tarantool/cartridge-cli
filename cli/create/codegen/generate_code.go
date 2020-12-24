// +build ignore

package main

import (
	"fmt"

	"github.com/shurcooL/vfsgen"
	"github.com/tarantool/cartridge-cli/cli/create/codegen/static"
)

func main() {
	err := vfsgen.Generate(static.CartridgeData, vfsgen.Options{
		PackageName:  "static",
		BuildTags:    "!dev",
		VariableName: "CartridgeData",
	})

	if err != nil {
		fmt.Errorf("Error while generating static files assets: %s", err)
	}
}
