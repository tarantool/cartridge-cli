package templates

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/shurcooL/httpfs/vfsutil"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
	"github.com/tarantool/cartridge-cli/cli/create/codegen/static"
	"github.com/tarantool/cartridge-cli/cli/templates"
)

const (
	CartridgeTemplateName = "cartridge"
)

var (
	knownTemplates = map[string]*templates.FileTreeTemplate{}
)

// Instantiate creates a file tree in a ctx.Project.Path according to ctx.Project.Template
// It applies ctx.Project to the template
func Instantiate(ctx *context.Ctx) error {
	var err error
	var projectTmpl *templates.FileTreeTemplate

	if ctx.Create.From != "" {
		log.Debugf("Template from %s is used", ctx.Create.From)

		if fileInfo, err := os.Stat(ctx.Create.From); err != nil {
			return fmt.Errorf("Failed to use specified path: %s", err)
		} else if !fileInfo.IsDir() {
			return fmt.Errorf("Specified path is not a directory: %s", ctx.Create.From)
		}

		// check specified template
		rocksPath := filepath.Join(ctx.Create.From, ".rocks")
		if _, err := os.Stat(rocksPath); !os.IsNotExist(err) {
			return fmt.Errorf(
				"Project template shouldn't contain .rocks directory. " +
					"To specify dependencies use rockspec and cartridge.pre-build hook",
			)
		}

		gitPath := filepath.Join(ctx.Create.From, ".git")
		if _, err := os.Stat(gitPath); !os.IsNotExist(err) {
			log.Warnf(
				"Project template contains .git directory. " +
					"It will be ignored on template instantiating",
			)
		}

		projectTmpl, err = parseTemplate(ctx.Create.From)
		if err != nil {
			return fmt.Errorf("Failed to parse template from specified path: %w", err)
		}
	} else {
		projectTmpl, err = parseStaticTemplate(ctx.Create.TemplateFS)

		if err != nil {
			return fmt.Errorf("Failed to parse template: %w", err)
		}
	}

	if err := projectTmpl.Instantiate(ctx.Project.Path, ctx.Project); err != nil {
		return fmt.Errorf("Failed to instantiate project template: %s", err)
	}

	return nil
}

func parseStaticTemplate(fs http.FileSystem) (*templates.FileTreeTemplate, error) {
	var tmpl templates.FileTreeTemplate

	err := vfsutil.Walk(fs, "/", func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
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
				Mode:    os.FileMode(static.FileModes[filePath]),
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

func parseTemplate(from string) (*templates.FileTreeTemplate, error) {
	var tmpl templates.FileTreeTemplate

	err := filepath.Walk(from, func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(from, filePath)
		if err != nil {
			return fmt.Errorf("Failed to get file path relative to the project root: %s", err)
		}

		// skip .git folder
		if relPath == "git" || strings.HasPrefix(relPath, ".git/") {
			return nil
		}

		if fileInfo.IsDir() {
			tmpl.AddDirs(templates.DirTemplate{
				Path: relPath,
				Mode: fileInfo.Mode(),
			})
		} else {
			fileContent, err := common.GetFileContent(filePath)
			if err != nil {
				return fmt.Errorf("Failed to get file content: %s", err)
			}

			tmpl.AddFiles(templates.FileTemplate{
				Path:    relPath,
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
