package templates

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/tarantool/cartridge-cli/project"
)

const (
	CartridgeTemplateName = "cartridge"
)

type fileTemplate struct {
	Path    string
	Mode    os.FileMode
	Content string
}

type dirTemplate struct {
	Path string
	Mode os.FileMode
}

type projectTemplate struct {
	Files []fileTemplate
	Dirs  []dirTemplate
}

var (
	knownTemplates = map[string]*projectTemplate{}
)

func init() {
	knownTemplates[CartridgeTemplateName] = combineTemplates(
		appFilesTemplate,
		buildFilesTemplate,
		configFilesTemplate,
		devFilesTemplate,
		testFilesTemplate,
	)
}

func combineTemplates(tmplts ...projectTemplate) *projectTemplate {
	var res projectTemplate

	for _, t := range tmplts {
		res.Files = append(res.Files, t.Files...)
		res.Dirs = append(res.Dirs, t.Dirs...)
	}

	return &res
}

func Instantiate(projectCtx *project.ProjectCtx) error {
	projectTmpl, exists := knownTemplates[projectCtx.Template]
	if !exists {
		return fmt.Errorf("Template %s does not exists", projectCtx.Template)
	}

	if err := createTree(projectTmpl, projectCtx); err != nil {
		return fmt.Errorf("Failed to instantiate %s template: %s", projectCtx.Template, err)
	}

	return nil
}

func createTree(tmpl *projectTemplate, projectCtx *project.ProjectCtx) error {
	// create dirs
	for _, d := range tmpl.Dirs {
		if err := createDir(&d, projectCtx); err != nil {
			return fmt.Errorf("Failed to create directory %s: %s", d.Path, err)
		}
	}

	// create files
	for _, t := range tmpl.Files {
		err := createFile(&t, projectCtx)
		if err != nil {
			return fmt.Errorf("Failed to create file %s: %s", t.Path, err)
		}
	}

	return nil
}

func createFile(t *fileTemplate, projectCtx *project.ProjectCtx) error {
	var err error

	// get a file path
	filePath, err := getTemplatedStr(&t.Path, projectCtx)
	if err != nil {
		return fmt.Errorf("Failed to get file path by template: %s", t.Path)
	}

	// create a file
	fullFilePath := filepath.Join(projectCtx.Path, *filePath)
	f, err := os.OpenFile(fullFilePath, os.O_CREATE|os.O_WRONLY, t.Mode)
	if err != nil {
		return err
	}

	defer f.Close()

	// write templated content to file
	fileContentTmpl, err := template.New("content").Parse(t.Content)
	if err != nil {
		return fmt.Errorf("Failed to parse a file content template: %s", t.Path)
	}

	if err := fileContentTmpl.Execute(f, projectCtx); err != nil {
		return fmt.Errorf("Failed to template a file content: %s", t.Path)
	}

	return nil
}

func createDir(d *dirTemplate, projectCtx *project.ProjectCtx) error {
	// get a dir path
	dirPath, err := getTemplatedStr(&d.Path, projectCtx)
	if err != nil {
		return fmt.Errorf("Failed to get dir path by template: %s", d.Path)
	}

	// create dir
	fullDirPath := filepath.Join(projectCtx.Path, *dirPath)
	if err := os.MkdirAll(fullDirPath, d.Mode); err != nil {
		return fmt.Errorf("Failed to create directory %s: %s", d.Path, err)
	}

	return nil
}

func getTemplatedStr(text *string, obj interface{}) (*string, error) {
	tmpl, err := template.New("path").Parse(*text)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err = tmpl.Execute(buf, obj); err != nil {
		return nil, err
	}

	res := buf.String()

	return &res, nil
}
