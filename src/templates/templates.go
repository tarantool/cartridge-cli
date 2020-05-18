package templates

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

type FileTemplate struct {
	Path    string
	Mode    os.FileMode
	Content string
}

type DirTemplate struct {
	Path string
	Mode os.FileMode
}

type FileTreeTemplate struct {
	Files []FileTemplate
	Dirs  []DirTemplate
}

func Combine(tmplts ...FileTreeTemplate) *FileTreeTemplate {
	var res FileTreeTemplate

	for _, t := range tmplts {
		res.Files = append(res.Files, t.Files...)
		res.Dirs = append(res.Dirs, t.Dirs...)
	}

	return &res
}

func InstantiateTree(tmpl *FileTreeTemplate, destDir string, ctx interface{}) error {
	// create dirs
	for _, d := range tmpl.Dirs {
		if err := InstantiateDir(&d, destDir, ctx); err != nil {
			return fmt.Errorf("Failed to create directory %s: %s", d.Path, err)
		}
	}

	// create files
	for _, t := range tmpl.Files {
		err := InstantiateFile(&t, destDir, ctx)
		if err != nil {
			return fmt.Errorf("Failed to create file %s: %s", t.Path, err)
		}
	}

	return nil
}

func InstantiateFile(t *FileTemplate, destDir string, ctx interface{}) error {
	var err error

	// get a file path
	filePath, err := getTemplatedStr(&t.Path, ctx)
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
		return fmt.Errorf("Failed to template a file %s content", t.Path, err)
	}

	return nil
}

func InstantiateDir(d *DirTemplate, destDir string, ctx interface{}) error {
	// get a dir path
	dirPath, err := getTemplatedStr(&d.Path, ctx)
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

func getTemplatedStr(text *string, obj interface{}) (string, error) {
	tmpl, err := template.New("path").Parse(*text)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	if err = tmpl.Execute(buf, obj); err != nil {
		return "", err
	}

	return buf.String(), nil
}
