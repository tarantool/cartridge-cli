package common

import (
	"fmt"

	"github.com/fatih/color"
)

type ResStatusType int

const (
	ResStatusOk ResStatusType = iota
	ResStatusSkipped
	ResStatusFailed
	ResStatusExited
)

type Result struct {
	ID     string
	Status ResStatusType
	Error  error
}

type ResChan chan Result

func (res *Result) String() string {
	resString, found := resStrings[res.Status]
	if !found {
		resString = fmt.Sprintf("Status %d", res.Status)
	}

	return fmt.Sprintf("%s... %s", res.ID, resString)
}

func (res *Result) FormatError() error {
	return fmt.Errorf("%s: %s", res.ID, res.Error)
}

var (
	ColorErr     *color.Color
	ColorWarn    *color.Color
	ColorOk      *color.Color
	ColorCyan    *color.Color
	ColorMagenta *color.Color

	resStrings map[ResStatusType]string
)

func init() {
	ColorErr = color.New(color.FgRed)
	ColorWarn = color.New(color.FgYellow)
	ColorOk = color.New(color.FgGreen)
	ColorCyan = color.New(color.FgCyan)
	ColorMagenta = color.New(color.FgHiMagenta)

	// resStrings
	resStrings = make(map[ResStatusType]string)
	resStrings[ResStatusOk] = ColorOk.Sprintf("OK")
	resStrings[ResStatusSkipped] = ColorWarn.Sprintf("SKIPPED")
	resStrings[ResStatusFailed] = ColorErr.Sprintf("FAILED")
	resStrings[ResStatusExited] = ColorErr.Sprintf("EXITED")
}
