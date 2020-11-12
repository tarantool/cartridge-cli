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
	ResStatusUpdated
	ResStatusCreated
)

type ResMessageType int

const (
	ResMessageWarn ResMessageType = iota + 10
	ResMessageDebug
	ResMessageInfo
	ResMessageErr
)

type ResultMessage struct {
	Type ResMessageType
	Text string
}

type Result struct {
	ID       string
	Status   ResStatusType
	Error    error
	Messages []ResultMessage
}

type ResChan chan Result

func GetMessage(msgType ResMessageType, format string, a ...interface{}) ResultMessage {
	return ResultMessage{
		Type: msgType,
		Text: fmt.Sprintf(format, a...),
	}
}

func GetWarnMessage(format string, a ...interface{}) ResultMessage {
	return GetMessage(ResMessageWarn, format, a...)
}

func GetDebugMessage(format string, a ...interface{}) ResultMessage {
	return GetMessage(ResMessageDebug, format, a...)
}

func GetInfoMessage(format string, a ...interface{}) ResultMessage {
	return GetMessage(ResMessageInfo, format, a...)
}

func GetErrMessage(format string, a ...interface{}) ResultMessage {
	return GetMessage(ResMessageErr, format, a...)
}

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
	ColorErr  *color.Color
	ColorWarn *color.Color
	ColorOk   *color.Color

	ColorRed     *color.Color
	ColorYellow  *color.Color
	ColorGreen   *color.Color
	ColorCyan    *color.Color
	ColorMagenta *color.Color

	resStrings map[ResStatusType]string
)

func init() {
	ColorRed = color.New(color.FgRed)
	ColorYellow = color.New(color.FgYellow)
	ColorGreen = color.New(color.FgGreen)
	ColorCyan = color.New(color.FgCyan)
	ColorMagenta = color.New(color.FgHiMagenta)

	ColorErr = ColorRed
	ColorWarn = ColorYellow
	ColorOk = ColorGreen

	// resStrings
	resStrings = make(map[ResStatusType]string)
	resStrings[ResStatusOk] = ColorOk.Sprintf("OK")
	resStrings[ResStatusSkipped] = ColorWarn.Sprintf("SKIPPED")
	resStrings[ResStatusFailed] = ColorErr.Sprintf("FAILED")
	resStrings[ResStatusExited] = ColorErr.Sprintf("EXITED")
	resStrings[ResStatusCreated] = ColorYellow.Sprintf("CREATED")
	resStrings[ResStatusUpdated] = ColorYellow.Sprintf("UPDATED")
}
