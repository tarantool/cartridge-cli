package running

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/fatih/color"
	psutil "github.com/shirou/gopsutil/process"
	log "github.com/sirupsen/logrus"
	"github.com/tarantool/cartridge-cli/cli/project"
)

type ProcStatusType int

const (
	procError ProcStatusType = iota
	procNotStarted
	procRunning
	procStopped
)

var (
	statusStrings map[ProcStatusType]string
)

func init() {
	// statusStrings
	statusStrings = make(map[ProcStatusType]string)
	statusStrings[procError] = color.New(color.FgRed).Sprintf("ERROR")
	statusStrings[procNotStarted] = color.New(color.FgCyan).Sprintf("NOT STARTED")
	statusStrings[procRunning] = color.New(color.FgGreen).Sprintf("RUNNING")
	statusStrings[procStopped] = color.New(color.FgYellow).Sprintf("STOPPED")
}

func getStatusStr(process *Process) string {
	statusStr, found := statusStrings[process.Status]
	if !found {
		return fmt.Sprintf("Status %d", process.Status)
	}

	return fmt.Sprintf("%s: %s", process.ID, statusStr)
}

type Process struct {
	ID     string
	Status ProcStatusType
	Error  error

	entrypoint string

	runDir  string
	workDir string
	pidFile string
	logDir  string
	logFile string

	env []string

	cmd       *exec.Cmd
	pid       int
	osProcess *psutil.Process
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

func (process *Process) Start(daemonize bool) error {
	var err error

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

	// initialize logs writer
	if !daemonize {
		logsWriter, err := newColorizedWriter(process)
		if err != nil {
			return fmt.Errorf("Failed to create colorized logs writer: %s", err)
		}

		process.cmd.Stdout = logsWriter
		process.cmd.Stderr = logsWriter
	} else {
		// create logs dir
		if err := os.MkdirAll(process.logDir, 0755); err != nil {
			return fmt.Errorf("Failed to initialize logs dir: %s", err)
		}

		// create logs file
		logFile, err := os.Create(process.logFile)
		if err != nil {
			return fmt.Errorf("Failed to create instance log file: %s", err)
		}
		defer logFile.Close()

		process.cmd.Stdout = logFile
		process.cmd.Stderr = logFile
	}

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

func NewInstanceProcess(projectCtx *project.ProjectCtx, instanceName string) *Process {
	var process Process

	process.ID = fmt.Sprintf("%s.%s", projectCtx.Name, instanceName)

	process.entrypoint = filepath.Join(projectCtx.Path, projectCtx.Entrypoint)
	process.runDir = projectCtx.RunDir
	process.pidFile = project.GetInstancePidFile(projectCtx, instanceName)
	process.workDir = project.GetInstanceWorkDir(projectCtx, instanceName)
	process.logDir = projectCtx.LogDir
	process.logFile = project.GetInstanceLogFile(projectCtx, instanceName)
	consoleSock := project.GetInstanceConsoleSock(projectCtx, instanceName)

	process.env = append(process.env,
		formatEnv("TARANTOOL_APP_NAME", projectCtx.Name),
		formatEnv("TARANTOOL_INSTANCE_NAME", instanceName),
		formatEnv("TARANTOOL_CFG", projectCtx.ConfPath),
		formatEnv("TARANTOOL_CONSOLE_SOCK", consoleSock),
		formatEnv("TARANTOOL_PID_FILE", process.pidFile),
		formatEnv("TARANTOOL_WORKDIR", process.workDir),
	)

	process.SetPidAndStatus()

	return &process
}

func NewStateboardProcess(projectCtx *project.ProjectCtx) *Process {
	var process Process

	process.ID = projectCtx.StateboardName

	process.entrypoint = filepath.Join(projectCtx.Path, projectCtx.StateboardEntrypoint)
	process.runDir = projectCtx.RunDir
	process.pidFile = project.GetStateboardPidFile(projectCtx)
	process.workDir = project.GetStateboardWorkDir(projectCtx)
	process.logDir = projectCtx.LogDir
	process.logFile = project.GetStateboardLogFile(projectCtx)
	consoleSock := project.GetStateboardConsoleSock(projectCtx)

	process.env = append(process.env,
		formatEnv("TARANTOOL_APP_NAME", projectCtx.StateboardName),
		formatEnv("TARANTOOL_CFG", projectCtx.ConfPath),
		formatEnv("TARANTOOL_CONSOLE_SOCK", consoleSock),
		formatEnv("TARANTOOL_PID_FILE", process.pidFile),
		formatEnv("TARANTOOL_WORKDIR", process.workDir),
	)

	process.SetPidAndStatus()

	return &process
}
