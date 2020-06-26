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

	Verbose bool
}

// FillCtx fills project context
func FillCtx(projectCtx *ProjectCtx) error {
	projectCtx.StateboardName = fmt.Sprintf("%s-stateboard", projectCtx.Name)

	return nil
}
