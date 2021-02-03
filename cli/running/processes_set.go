package running

import (
	"fmt"
	"time"

	"github.com/apex/log"
	"github.com/tarantool/cartridge-cli/cli/common"
)

type ProcessesSet []*Process

func (set *ProcessesSet) Add(processes ...*Process) {
	*set = append(*set, processes...)
}

func startProcess(process *Process, daemonize bool, disableLogPrefix bool, timeout time.Duration, resCh common.ResChan) {
	if process.Status == procStatusError {
		resCh <- common.Result{
			ID:     process.ID,
			Status: common.ResStatusFailed,
			Error:  process.Error,
		}
		return
	}

	if process.Status == procStatusRunning {
		resCh <- common.Result{
			ID:     process.ID,
			Status: common.ResStatusSkipped,
			Error:  fmt.Errorf("Process is already running"),
		}
		return
	}

	if err := process.Start(daemonize, disableLogPrefix); err != nil {
		resCh <- common.Result{
			ID:     process.ID,
			Status: common.ResStatusFailed,
			Error:  fmt.Errorf("Failed to start: %s", err),
		}
		return
	}

	if daemonize {
		if err := process.WaitReady(timeout); err != nil {
			resCh <- common.Result{
				ID:     process.ID,
				Status: common.ResStatusFailed,
				Error:  fmt.Errorf("Failed to wait process is ready: %s", err),
			}
			return
		}

		resCh <- common.Result{
			ID:     process.ID,
			Status: common.ResStatusOk,
		}
		return
	}

	if err := process.Wait(); err != nil {
		resCh <- common.Result{
			ID:     process.ID,
			Status: common.ResStatusExited,
			Error:  fmt.Errorf("Process exited: %s", err),
		}
	} else {
		resCh <- common.Result{
			ID:     process.ID,
			Status: common.ResStatusExited,
		}
	}
}

func (set *ProcessesSet) Start(daemonize bool, disableLogPrefix bool, timeout time.Duration) error {
	resCh := make(chan common.Result)

	for _, process := range *set {
		go startProcess(process, daemonize, disableLogPrefix, timeout, resCh)

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
			log.Infof(res.String())
			if res.Error != nil {
				if !daemonize {
					log.Errorf("%s: %s", res.ID, res.Error)
				} else {
					errors = append(errors, res.FormatError())
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

func (set *ProcessesSet) Stop(force bool) error {
	var errors []error
	var warnings []error

	for _, process := range *set {
		var res common.Result

		if process.Status == procStatusError {
			res = common.Result{
				ID:     process.ID,
				Status: common.ResStatusFailed,
				Error:  process.Error,
			}
		} else if process.Status == procStatusStopped || process.Status == procStatusNotStarted {
			res = common.Result{
				ID:     process.ID,
				Status: common.ResStatusSkipped,
				Error:  fmt.Errorf("Process is not running"),
			}
		} else if err := process.Stop(force); err != nil {
			res = common.Result{
				ID:     process.ID,
				Status: common.ResStatusFailed,
				Error:  fmt.Errorf("Failed to stop: %s", err),
			}
		} else {
			res = common.Result{
				ID:     process.ID,
				Status: common.ResStatusOk,
			}
		}

		if res.Status == common.ResStatusFailed {
			errors = append(errors, res.FormatError())
		}

		if res.Status == common.ResStatusSkipped {
			warnings = append(warnings, res.FormatError())
		}

		log.Infof(res.String())
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

func clearProcessData(process *Process, resCh common.ResChan) {
	if process.Status == procStatusError {
		resCh <- common.Result{
			ID:     process.ID,
			Status: common.ResStatusFailed,
			Error:  process.Error,
		}
		return
	}
	if process.Status == procStatusRunning {
		resCh <- common.Result{
			ID:     process.ID,
			Status: common.ResStatusFailed,
			Error:  fmt.Errorf("Instance is running"),
		}
		return
	}

	resCh <- process.Clean()
}

func (set *ProcessesSet) Clean() error {
	resCh := make(common.ResChan)

	for _, process := range *set {
		go clearProcessData(process, resCh)
	}

	var errors []error
	var warnings []error

	// wait for all processes result
	for i := 0; i < len(*set); i++ {
		select {
		case res := <-resCh:
			if res.Status == common.ResStatusFailed {
				errors = append(errors, res.FormatError())
			}

			if res.Status == common.ResStatusSkipped || res.Status == common.ResStatusOk && res.Error != nil {
				warnings = append(warnings, res.FormatError())
			}

			log.Infof(res.String())
		}
	}

	if len(warnings) > 0 {
		for _, warn := range warnings {
			log.Debugf("%s", warn)
		}
	}

	if len(errors) > 0 {
		for _, err := range errors {
			log.Errorf("%s", err)
		}
		return fmt.Errorf("Failed to clean some instances data")
	}

	return nil
}

func getProcessLogs(process *Process, follow bool, n int, resCh common.ResChan) {
	if process.Status == procStatusError {
		resCh <- common.Result{
			ID:     process.ID,
			Status: common.ResStatusFailed,
			Error:  process.Error,
		}
	} else if err := process.Log(follow, n); err != nil {
		resCh <- common.Result{
			ID:     process.ID,
			Status: common.ResStatusFailed,
			Error:  fmt.Errorf("Failed to get logs: %s", err),
		}
	} else {
		resCh <- common.Result{
			ID:     process.ID,
			Status: common.ResStatusOk,
		}
	}
}

func (set *ProcessesSet) Log(follow bool, lines int) error {
	resCh := make(chan common.Result)

	for _, process := range *set {
		go getProcessLogs(process, follow, lines, resCh)

		// wait for process to print logs
		time.Sleep(100 * time.Millisecond)
	}

	var errors []error

	// wait for all processes result
	for i := 0; i < len(*set); i++ {
		select {
		case res := <-resCh:
			log.Infof(res.String())
			if res.Error != nil {
				if follow {
					log.Errorf("%s: %s", res.ID, res.Error)
				} else {
					errors = append(errors, res.FormatError())
				}
			}
		}
	}

	if len(errors) > 0 {
		for _, err := range errors {
			log.Errorf("%s", err)
		}
		return fmt.Errorf("Failed to get some instances logs")
	}

	return nil
}
