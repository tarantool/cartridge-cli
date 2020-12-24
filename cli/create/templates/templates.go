package templates

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/shurcooL/httpfs/vfsutil"

	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/create/codegen/static"
	"github.com/tarantool/cartridge-cli/cli/templates"
)

const (
	CartridgeTemplateName = "cartridge"
)

var (
	knownTemplates = map[string]*templates.FileTreeTemplate{}
	cartridgeFs    = static.CartridgeData
)

func init() {
	/*
		knownTemplates[CartridgeTemplateName] = templates.Combine(
			appFilesTemplate,
			buildFilesTemplate,
			configFilesTemplate,
			devFilesTemplate,
			testFilesTemplate,
		)
	*/
}

// Instantiate creates a file tree in a ctx.Project.Path according to ctx.Project.Template
// It applies ctx.Project to the template
func Instantiate(ctx *context.Ctx) error {
	var err error
	var projectTmpl *templates.FileTreeTemplate

	if ctx.Create.FileSystem != nil {
		log.Debugf("Template from %s is used", ctx.Create.FileSystem)

		projectTmpl, err = parseTemplate(ctx.Create.FileSystem)

		if err != nil {
			return fmt.Errorf("Failed to parse template from specified path: %w", err)
		}
	} else {
		return fmt.Errorf("Failed ")
	}

	if err := projectTmpl.Instantiate(ctx.Project.Path, ctx.Project); err != nil {
		return fmt.Errorf("Failed to instantiate project template: %s", err)
	}

	return nil
}

func parseTemplate(fs http.FileSystem) (*templates.FileTreeTemplate, error) {
	var tmpl templates.FileTreeTemplate

	err := vfsutil.Walk(fs, "/", func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip .git folder
		if filePath == "git" || strings.HasPrefix(filePath, ".git/") {
			log.Warnf(
				"Project template contains .git directory. " +
					"It will be ignored on template instantiating",
			)

			return nil
		}

		if fileInfo.IsDir() {
			tmpl.AddDirs(templates.DirTemplate{
				Path: filePath,
				Mode: fileInfo.Mode(),
			})
		} else {
			fileContent, err := getStaticFileContent(fs, filePath)
			if err != nil {
				return fmt.Errorf("Failed to get file content: %s", err)
			}

			tmpl.AddFiles(templates.FileTemplate{
				Path:    filePath,
				Mode:    fileInfo.Mode(),
				Content: fileContent,
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to parse template: %s", err)
	}

	return &tmpl, nil
}

// getStaticFileContent open file in generated static filesystem
func getStaticFileContent(fs http.FileSystem, filename string) (string, error) {
	file, err := fs.Open(filename)
	if err != nil {
		log.Errorf("Failed to open static file: %s", err)
		return "", err
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		log.Errorf("Failed to get static file content: %s", err)
		return "", err
	}

	defer file.Close()

	return string(content), nil
}
