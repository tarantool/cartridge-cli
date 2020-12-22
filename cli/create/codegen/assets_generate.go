// +build ignore

package main

import (
	"log"

	"github.com/shurcooL/vfsgen"
	"github.com/tarantool/cartridge-cli/cli/create/static"
)

func main() {
	err := vfsgen.Generate(static.Data, vfsgen.Options{
		PackageName:  "static",
		BuildTags:    "!dev",
		VariableName: "Data",
	})

	if err != nil {
		log.Fatalln(err)
	}
}
