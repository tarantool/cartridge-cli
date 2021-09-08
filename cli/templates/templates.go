package templates

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// FileTemplate stores a single file template
type FileTemplate struct {
	Path    string
	Mode    os.FileMode
	Content string
}

// DirTemplate stores a single directory template
type DirTemplate struct {
	Path string
	Mode os.FileMode
}

// FileTreeTemplate stores filetree template
type FileTreeTemplate struct {
	Files []FileTemplate
	Dirs  []DirTemplate
}

// Template is the interface that has Instantiate method
type Template interface {
	Instantiate(destDir string, ctx interface{}) error
}

// AddFiles adds files to tree template
func (tmpl *FileTreeTemplate) AddFiles(fileTemplates ...FileTemplate) {
	for _, f := range fileTemplates {
		tmpl.Files = append(tmpl.Files, f)
	}
}

// AddDirs adds dirs to tree template
func (tmpl *FileTreeTemplate) AddDirs(dirTemplates ...DirTemplate) {
	for _, d := range dirTemplates {
		tmpl.Dirs = append(tmpl.Dirs, d)
	}
}

// Combine combines two file tree templates
func Combine(tmplts ...FileTreeTemplate) *FileTreeTemplate {
	var res FileTreeTemplate

	for _, t := range tmplts {
		res.Files = append(res.Files, t.Files...)
		res.Dirs = append(res.Dirs, t.Dirs...)
	}

	return &res
}

// Instantiate instantiates file tree template
func (tmpl *FileTreeTemplate) Instantiate(destDir string, ctx interface{}) error {
	// create dirs
	for _, d := range tmpl.Dirs {
		if err := d.Instantiate(destDir, ctx); err != nil {
			return fmt.Errorf("Failed to create directory %s: %s", d.Path, err)
		}
	}

	// create files
	for _, t := range tmpl.Files {
		err := t.Instantiate(destDir, ctx)
		if err != nil {
			return fmt.Errorf("Failed to create file %s: %s", t.Path, err)
		}
	}

	return nil
}

// Instantiate instantiates file template
func (t *FileTemplate) Instantiate(destDir string, ctx interface{}) error {
	var err error

	// get a file path
	filePath, err := GetTemplatedStr(&t.Path, ctx)
	if err != nil {
		return fmt.Errorf("Failed to get file path by template: %s", t.Path)
	}

	// create a file
	fullFilePath := filepath.Join(destDir, filePath)
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

	if err := fileContentTmpl.Execute(f, ctx); err != nil {
		return fmt.Errorf("Failed to template a file %s content: %s", t.Path, err)
	}

	return nil
}

// Instantiate instantiates file template
func (d *DirTemplate) Instantiate(destDir string, ctx interface{}) error {
	// get a dir path
	dirPath, err := GetTemplatedStr(&d.Path, ctx)
	if err != nil {
		return fmt.Errorf("Failed to get dir path by template: %s", d.Path)
	}

	// create dir
	fullDirPath := filepath.Join(destDir, dirPath)
	if err := os.MkdirAll(fullDirPath, d.Mode); err != nil {
		return fmt.Errorf("Failed to create directory %s: %s", d.Path, err)
	}

	return nil
}

func GetTemplatedStr(text *string, obj interface{}) (string, error) {
	// If the user specifies the name of the application with capital letters
	// (`cartridge create` command), we have to create rockspec file in lowercase,
	// otherwise an error will occur at the build stage. For this we use Funcs,
	// embedding ToLower there. See https://github.com/tarantool/cartridge-cli/issues/610
	// for more details.

	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
	}

	tmpl, err := template.New("s").Funcs(funcMap).Parse(*text)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err = tmpl.Execute(buf, obj); err != nil {
		return "", err
	}

	return buf.String(), nil
}
