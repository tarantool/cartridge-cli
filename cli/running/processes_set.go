package running

import (
	"fmt"
	"time"

	"github.com/apex/log"
	"github.com/fatih/color"
)

type ProcessesSet []*Process

const (
	procResOk procResType = iota
	procResSkipped
	procResFailed
	procResExited
)

type procResType int
type ProcessRes struct {
	ProcessID string
	Res       procResType
	Error     error
}

var (
	resStrings map[procResType]string
)

func init() {
	// resStrings
	resStrings = make(map[procResType]string)
	resStrings[procResOk] = color.New(color.FgGreen).Sprintf("OK")
	resStrings[procResSkipped] = color.New(color.FgYellow).Sprintf("SKIPPED")
	resStrings[procResFailed] = color.New(color.FgRed).Sprintf("FAILED")
	resStrings[procResExited] = color.New(color.FgRed).Sprintf("EXITED")
}

func getResStr(processRes *ProcessRes) string {
	resString, found := resStrings[processRes.Res]
	if !found {
		resString = fmt.Sprintf("Status %d", processRes.Res)
	}

	return fmt.Sprintf("%s... %s", processRes.ProcessID, resString)
}

func (set *ProcessesSet) Add(processes ...*Process) {
	*set = append(*set, processes...)
}

func startProcess(process *Process, daemonize bool, resCh chan ProcessRes) {
	if process.Status == procStatusError {
		resCh <- ProcessRes{
			ProcessID: process.ID,
			Res:       procResFailed,
			Error:     process.Error,
		}
		return
	}

	if process.Status == procStatusRunning {
		resCh <- ProcessRes{
			ProcessID: process.ID,
			Res:       procResSkipped,
			Error:     fmt.Errorf("Process is already running"),
		}
		return
	}

	if err := process.Start(daemonize); err != nil {
		resCh <- ProcessRes{
			ProcessID: process.ID,
			Res:       procResFailed,
			Error:     fmt.Errorf("Failed to start: %s", err),
		}
		return
	}

	if daemonize {
		if err := process.WaitReady(); err != nil {
			resCh <- ProcessRes{
				ProcessID: process.ID,
				Res:       procResFailed,
				Error:     fmt.Errorf("Failed to wait process is ready: %s", err),
			}
			return
		}

		resCh <- ProcessRes{
			ProcessID: process.ID,
			Res:       procResOk,
		}
		return
	}

	if err := process.Wait(); err != nil {
		resCh <- ProcessRes{
			ProcessID: process.ID,
			Res:       procResExited,
			Error:     fmt.Errorf("Process exited: %s", err),
		}
	} else {
		resCh <- ProcessRes{
			ProcessID: process.ID,
			Res:       procResExited,
		}
	}
}

func (set *ProcessesSet) Start(daemonize bool) error {
	resCh := make(chan ProcessRes)

	for _, process := range *set {
		go startProcess(process, daemonize, resCh)

		// wait for process to print logs
		if !daemonize {
			time.Sleep(200 * time.Millisecond)
		}
	}

	var errors []error

	// wait for all processes result
	for i := 0; i < len(*set); i++ {
		select {
		case res := <-resCh:
			log.Infof(getResStr(&res))
			if res.Error != nil {
				if !daemonize {
					log.Errorf("%s: %s", res.ProcessID, res.Error)
				} else {
					errors = append(errors, fmt.Errorf("%s: %s", res.ProcessID, res.Error))
				}
			}
		}
	}

	if !daemonize {
		return fmt.Errorf("All instances exited")
	}

	if len(errors) > 0 {
		for _, err := range errors {
			log.Errorf("%s", err)
		}
		return fmt.Errorf("Failed to start some instances")
	}

	return nil
}

func (set *ProcessesSet) Stop() error {
	var errors []error
	var warnings []error

	for _, process := range *set {
		var res ProcessRes

		if process.Status == procStatusError {
			res = ProcessRes{
				ProcessID: process.ID,
				Res:       procResFailed,
				Error:     process.Error,
			}
		} else if process.Status == procStatusStopped || process.Status == procStatusNotStarted {
			res = ProcessRes{
				ProcessID: process.ID,
				Res:       procResSkipped,
				Error:     fmt.Errorf("Process is not running"),
			}
		} else if err := process.Stop(); err != nil {
			res = ProcessRes{
				ProcessID: process.ID,
				Res:       procResFailed,
				Error:     fmt.Errorf("Failed to stop: %s", err),
			}
		} else {
			res = ProcessRes{
				ProcessID: process.ID,
				Res:       procResOk,
			}
		}

		if res.Res == procResFailed {
			errors = append(errors, fmt.Errorf("%s: %s", res.ProcessID, res.Error))
		}

		if res.Res == procResSkipped {
			warnings = append(warnings, fmt.Errorf("%s: %s", res.ProcessID, res.Error))
		}

		log.Infof(getResStr(&res))
	}

	if len(warnings) > 0 {
		for _, warn := range warnings {
			log.Warnf("%s", warn)
		}
	}

	if len(errors) > 0 {
		for _, err := range errors {
			log.Errorf("%s", err)
		}
		return fmt.Errorf("Failed to stop some instances")
	}

	return nil
}

func (set *ProcessesSet) Status() error {
	var errors []string

	for _, process := range *set {
		if process.Status == procStatusError {
			errors = append(errors, fmt.Sprintf("%s: %s", process.ID, process.Error))
		}

		log.Infof(getStatusStr(process))
	}

	if len(errors) > 0 {
		for _, err := range errors {
			log.Error(err)
		}
		return fmt.Errorf("Failed to get some instances status")
	}

	return nil
}
