package main

import (
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/dave/jennifer/jen"
)

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

func main() {
	err := generateFileModeFile(
		"cli/create/templates/cartridge",
		"cli/codegen/static/create_cartrdige_template_filemodes_gen.go",
	)

	if err != nil {
		log.Errorf("Error while generating file modes: %s", err)
	}
}
