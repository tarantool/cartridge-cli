package project

import (
	"fmt"
)

const (
	defaultTemplate = "cartridge"
)

type ProjectCtx struct {
	Name           string
	StateboardName string
	Path           string
	Template       string
}

func FillCtx(projectCtx *ProjectCtx) error {
	projectCtx.StateboardName = fmt.Sprintf("%s-stateboard", projectCtx.Name)

	return nil
}
