package running

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
)

type ProcessesSet []*Process

const (
	procOk procResType = iota + 10
	procSkipped
	procFailed
	procExited
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
	resStrings[procOk] = color.New(color.FgGreen).Sprintf("OK")
	resStrings[procSkipped] = color.New(color.FgYellow).Sprintf("SKIPPED")
	resStrings[procFailed] = color.New(color.FgRed).Sprintf("FAILED")
	resStrings[procExited] = color.New(color.FgRed).Sprintf("EXITED")
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

func (set *ProcessesSet) Start(daemonize bool) error {
	resCh := make(chan ProcessRes)

	for _, process := range *set {
		go func(process *Process) {
			if process.Status == procError {
				resCh <- ProcessRes{
					ProcessID: process.ID,
					Res:       procFailed,
					Error:     process.Error,
				}
				return
			}

			if process.Status == procRunning {
				resCh <- ProcessRes{
					ProcessID: process.ID,
					Res:       procSkipped,
					Error:     fmt.Errorf("Process is already running"),
				}
				return
			}

			if err := process.Start(daemonize); err != nil {
				resCh <- ProcessRes{
					ProcessID: process.ID,
					Res:       procFailed,
					Error:     fmt.Errorf("Failed to start: %s", err),
				}
				return
			}

			if daemonize {
				if err := process.WaitReady(); err != nil {
					resCh <- ProcessRes{
						ProcessID: process.ID,
						Res:       procFailed,
						Error:     fmt.Errorf("Failed to wait process is ready: %s", err),
					}
					return
				}

				resCh <- ProcessRes{
					ProcessID: process.ID,
					Res:       procOk,
				}
				return
			}

			if err := process.Wait(); err != nil {
				resCh <- ProcessRes{
					ProcessID: process.ID,
					Res:       procExited,
					Error:     fmt.Errorf("Process exited: %s", err),
				}
			} else {
				resCh <- ProcessRes{
					ProcessID: process.ID,
					Res:       procExited,
				}
			}
		}(process)

		if !daemonize {
			time.Sleep(200 * time.Millisecond)
		}
	}

	var errors []error

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
			log.Error(err)
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

		if process.Status == procError {
			res = ProcessRes{
				ProcessID: process.ID,
				Res:       procFailed,
				Error:     process.Error,
			}
		} else if process.Status == procStopped || process.Status == procNotStarted {
			res = ProcessRes{
				ProcessID: process.ID,
				Res:       procSkipped,
				Error:     fmt.Errorf("Process is not running"),
			}
		} else if err := process.Stop(); err != nil {
			res = ProcessRes{
				ProcessID: process.ID,
				Res:       procFailed,
				Error:     fmt.Errorf("Failed to stop: %s", err),
			}
		} else {
			res = ProcessRes{
				ProcessID: process.ID,
				Res:       procOk,
			}
		}

		if res.Res == procFailed {
			errors = append(errors, res.Error)
		}

		if res.Res == procSkipped {
			warnings = append(warnings, res.Error)
		}

		log.Infof(getResStr(&res))
	}

	if len(warnings) > 0 {
		for _, warn := range warnings {
			log.Warn(warn)
		}
	}

	if len(errors) > 0 {
		for _, err := range errors {
			log.Error(err)
		}
		return fmt.Errorf("Failed to stop some instances")
	}

	return nil
}

func (set *ProcessesSet) Status() error {
	var errors []string

	for _, process := range *set {
		if process.Status == procError {
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
