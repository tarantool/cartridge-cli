package running

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	psutil "github.com/shirou/gopsutil/process"
	log "github.com/sirupsen/logrus"
)

type procStatus int
type Process struct {
	ID     string
	Status procStatus
	Error  error

	entrypoint string
	runDir     string
	workDir    string
	pidFile    string
	env        []string

	cmd       *exec.Cmd
	pid       int
	osProcess *psutil.Process
	writer    io.Writer
}

type ProcessesSet []*Process

const (
	procError procStatus = iota
	procNotStarted
	procRunning
	procStopped

	procOk procRes = iota + 10
	procSkipped
	procFailed
	procExited
)

type procRes int
type ProcessRes struct {
	ProcessID string
	Res       procRes
	Error     error
}

var (
	statusStrings map[procStatus]string
	resStrings    map[procRes]string
)

func init() {
	// statusStrings
	statusStrings = make(map[procStatus]string)
	statusStrings[procError] = color.New(color.FgRed).Sprintf("ERROR")
	statusStrings[procNotStarted] = color.New(color.FgCyan).Sprintf("NOT STARTED")
	statusStrings[procRunning] = color.New(color.FgGreen).Sprintf("RUNNING")
	statusStrings[procStopped] = color.New(color.FgYellow).Sprintf("STOPPED")

	// resStrings
	resStrings = make(map[procRes]string)
	resStrings[procOk] = color.New(color.FgGreen).Sprintf("OK")
	resStrings[procSkipped] = color.New(color.FgYellow).Sprintf("SKIPPED")
	resStrings[procFailed] = color.New(color.FgRed).Sprintf("FAILED")
	resStrings[procExited] = color.New(color.FgRed).Sprintf("EXITED")
}

func getStatusStr(process *Process) string {
	statusStr, found := statusStrings[process.Status]
	if !found {
		return fmt.Sprintf("Status %d", process.Status)
	}

	return fmt.Sprintf("%s: %s", process.ID, statusStr)
}

func getResStr(processRes *ProcessRes) string {
	resString, found := resStrings[processRes.Res]
	if !found {
		resString = fmt.Sprintf("Status %d", processRes.Res)
	}

	return fmt.Sprintf("%s... %s", processRes.ProcessID, resString)
}

func (process *Process) SetPidAndStatus() {
	var err error

	pidFile, err := os.Open(process.pidFile)
	if os.IsNotExist(err) {
		process.Status = procNotStarted
		return
	}
	if err != nil {
		process.Status = procError
		process.Error = fmt.Errorf("Failed to check process PID file: %s", err)
		return
	}

	pidBytes, err := ioutil.ReadAll(pidFile)
	if err != nil {
		process.Status = procError
		process.Error = fmt.Errorf("Failed to read process PID from file: %s", err)
		return
	}

	process.pid, err = strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		process.Status = procError
		process.Error = fmt.Errorf("PID file exists with unknown format: %s", err)
		return
	}

	process.osProcess, err = psutil.NewProcess(int32(process.pid))
	if err != nil {
		process.Status = procStopped
		return
	}

	name, err := process.osProcess.Name()
	if err != nil {
		process.Status = procError
		process.Error = fmt.Errorf("Failed to get process %d name: %s", process.pid, name)
		return
	}

	if name != "tarantool" {
		process.Status = procError
		process.Error = fmt.Errorf("Process %d does not seem to be tarantool", process.pid)
		return
	}

	if err := process.osProcess.SendSignal(syscall.Signal(0)); err != nil {
		process.Status = procStopped
	} else {
		process.Status = procRunning
	}
}

func (process *Process) Start() error {
	ctx := context.Background()
	process.cmd = exec.CommandContext(ctx, "tarantool", process.entrypoint)

	process.cmd.Env = append(os.Environ(), process.env...)
	process.cmd.Dir = process.workDir

	// create run dir
	if err := os.MkdirAll(process.runDir, 0755); err != nil {
		return fmt.Errorf("Failed to initialize run dir: %s", err)
	}

	// create work dir
	if err := os.MkdirAll(process.workDir, 0755); err != nil {
		return fmt.Errorf("Failed to initialize work dir: %s", err)
	}

	// create pid file
	pidFile, err := os.Create(process.pidFile)
	if err != nil {
		return fmt.Errorf("Failed to create PID file: %s", err)
	}
	defer pidFile.Close()

	process.cmd.Stdout = process.writer
	process.cmd.Stderr = process.writer

	if err := process.cmd.Start(); err != nil {
		return fmt.Errorf("Failed to start: %s", err)
	}

	if _, err := pidFile.WriteString(strconv.Itoa(process.cmd.Process.Pid)); err != nil {
		log.Warnf("Failed to write PID %d: %s", process.cmd.Process.Pid, err)
	}

	return nil
}

func (process *Process) Wait() error {
	if err := process.cmd.Wait(); err != nil {
		return fmt.Errorf("Exited unsuccessfully: %s", err)
	}

	return nil
}

func (process *Process) Stop() error {
	if process.osProcess == nil {
		return fmt.Errorf("Process %d is not running", process.pid) // XXX: internal error
	}

	if err := process.osProcess.SendSignal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("Failed to terminate process %d: %s", process.pid, err)
	}

	return nil
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

			if err := process.Start(); err != nil {
				resCh <- ProcessRes{
					ProcessID: process.ID,
					Res:       procFailed,
					Error:     fmt.Errorf("Failed to start: %s", err),
				}
				return
			}

			if daemonize {
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
