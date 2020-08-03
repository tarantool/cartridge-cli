package running

import (
	goContext "context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/apex/log"
	"github.com/fatih/color"
	"github.com/hpcloud/tail"
	psutil "github.com/shirou/gopsutil/process"
	"github.com/tarantool/cartridge-cli/cli/context"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"
)

type ProcStatusType int

const (
	procStatusError ProcStatusType = iota
	procStatusNotStarted
	procStatusRunning
	procStatusStopped

	notifyReady   = "READY=1"
	notifyBufSize = 300
)

var (
	statusStrings      map[ProcStatusType]string
	notifyStatusRgx    *regexp.Regexp
	notifyRetryTimeout = 500 * time.Millisecond
)

func init() {
	// statusStrings
	statusStrings = make(map[ProcStatusType]string)
	statusStrings[procStatusError] = color.New(color.FgRed).Sprintf("ERROR")
	statusStrings[procStatusNotStarted] = color.New(color.FgCyan).Sprintf("NOT STARTED")
	statusStrings[procStatusRunning] = color.New(color.FgGreen).Sprintf("RUNNING")
	statusStrings[procStatusStopped] = color.New(color.FgYellow).Sprintf("STOPPED")

	notifyStatusRgx = regexp.MustCompile(`(?s:^STATUS=(.+)$)`)
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

	runDir      string
	workDir     string
	pidFile     string
	logDir      string
	logFile     string
	consoleSock string

	notifySockPath string
	notifyConn     net.PacketConn

	env []string

	cmd       *exec.Cmd
	pid       int
	osProcess *psutil.Process
}

func (process *Process) SetPidAndStatus() {
	var err error

	if process.pid == 0 {

		pidFile, err := os.Open(process.pidFile)
		if os.IsNotExist(err) {
			process.Status = procStatusNotStarted
			return
		}
		if err != nil {
			process.Status = procStatusError
			process.Error = fmt.Errorf("Failed to check process PID file: %s", err)
			return
		}

		pidBytes, err := ioutil.ReadAll(pidFile)
		if err != nil {
			process.Status = procStatusError
			process.Error = fmt.Errorf("Failed to read process PID from file: %s", err)
			return
		}

		if len(pidBytes) == 0 {
			process.Status = procStatusNotStarted
			return
		}

		process.pid, err = strconv.Atoi(strings.TrimSpace(string(pidBytes)))
		if err != nil {
			process.Status = procStatusError
			process.Error = fmt.Errorf("PID file exists with unknown format: %s", err)
			return
		}
	}

	process.osProcess, err = psutil.NewProcess(int32(process.pid))
	if err != nil {
		process.Status = procStatusStopped
		return
	}

	name, err := process.osProcess.Name()
	if err != nil {
		process.Status = procStatusError
		process.Error = fmt.Errorf("Failed to get process %d name: %s", process.pid, name)
		return
	}

	if name != "tarantool" {
		process.Status = procStatusError
		process.Error = fmt.Errorf("Process %d does not seem to be tarantool", process.pid)
		return
	}

	if err := process.osProcess.SendSignal(syscall.Signal(0)); err != nil {
		process.Status = procStatusStopped
	} else {
		process.Status = procStatusRunning
	}
}

func (process *Process) Start(daemonize bool) error {
	var err error

	if _, err := os.Stat(process.entrypoint); err != nil {
		return fmt.Errorf("Can't use instance entrypoint: %s", err)
	}

	// create run dir
	if err := os.MkdirAll(process.runDir, 0755); err != nil {
		return fmt.Errorf("Failed to initialize run dir: %s", err)
	}

	// create work dir
	if err := os.MkdirAll(process.workDir, 0755); err != nil {
		return fmt.Errorf("Failed to initialize work dir: %s", err)
	}

	if daemonize {
		if err := buildNotifySocket(process); err != nil {
			return fmt.Errorf("Failed to build notify socket: %s", err)
		}
	}

	ctx := goContext.Background()
	process.cmd = exec.CommandContext(ctx, "tarantool", process.entrypoint)

	process.cmd.Env = append(os.Environ(), process.env...)
	process.cmd.Dir = process.workDir

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

	// create pid file
	pidFile, err := os.Create(process.pidFile)
	if err != nil {
		return fmt.Errorf("Failed to create PID file: %s", err)
	}
	defer pidFile.Close()

	if err := process.cmd.Start(); err != nil {
		return fmt.Errorf("Failed to start: %s", err)
	}

	// save process PID
	process.pid = process.cmd.Process.Pid
	if _, err := pidFile.WriteString(strconv.Itoa(process.pid)); err != nil {
		log.Warnf("Failed to write PID %d: %s", process.pid, err)
	}

	return nil
}

func (process *Process) Wait() error {
	if err := process.cmd.Wait(); err != nil {
		return fmt.Errorf("Exited unsuccessfully: %s", err)
	}

	return nil
}

func (process *Process) WaitReady(timeout time.Duration) error {
	if process.notifyConn == nil {
		return fmt.Errorf("Notify connection wasn't created")
	}
	defer process.notifyConn.Close()

	timeStart := time.Now()

	for {
		if timeout != 0 && time.Now().Sub(timeStart) > timeout {
			log.Errorf("%s: Start timeout was reached. Killing the process...", process.ID)
			if err := process.Kill(); err != nil {
				log.Warnf("Failed to kill process %s: %s", process.ID, err)
			}
			return fmt.Errorf("Timeout was reached")
		}

		process.SetPidAndStatus()

		switch process.Status {
		case procStatusError:
			return fmt.Errorf("Failed to check process status: %s", process.Error)
		case procStatusNotStarted:
			return fmt.Errorf("Process isn't statred")
		case procStatusStopped:
			return fmt.Errorf("Process seems to be stopped")
		}

		if err := process.notifyConn.SetReadDeadline(time.Now().Add(notifyRetryTimeout)); err != nil {
			return fmt.Errorf("Failed to set read deadline for notify connection: %s", err)
		}

		buffer := make([]byte, notifyBufSize)
		if _, _, err := process.notifyConn.ReadFrom(buffer); err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			} else {
				return fmt.Errorf("Failed to read from notify socket: %s", err)
			}
		}

		msg := strings.TrimRight(string(buffer), "\x00")

		if msg == notifyReady {
			break
		}

		matches := notifyStatusRgx.FindStringSubmatch(msg)
		if matches == nil {
			return fmt.Errorf("Failed to parse notify message: %s", msg)
		}

		status := matches[1]
		if strings.HasPrefix(status, "Failed") {
			return fmt.Errorf("Failed to start: %s", status)
		}
	}

	return nil
}

func (process *Process) SendSignal(sig syscall.Signal) error {
	if process.osProcess == nil {
		return project.InternalError("Process %d is not running", process.pid)
	}

	if err := process.osProcess.SendSignal(sig); err != nil {
		return fmt.Errorf("Failed to send %s signal to process %d: %s", sig.String(), process.pid, err)
	}

	return nil
}

func (process *Process) Stop(force bool) error {
	if !force {
		return process.Terminate()
	}
	return process.Kill()
}

func (process *Process) Terminate() error {
	return process.SendSignal(syscall.SIGTERM)
}

func (process *Process) Kill() error {
	return process.SendSignal(syscall.SIGKILL)
}

func getEntrypointPath(appPath string, specifiedEntrypoint string) string {
	if filepath.IsAbs(specifiedEntrypoint) {
		return specifiedEntrypoint
	}

	return filepath.Join(appPath, specifiedEntrypoint)
}

func NewInstanceProcess(ctx *context.Ctx, instanceName string) *Process {
	var process Process

	process.ID = fmt.Sprintf("%s.%s", ctx.Project.Name, instanceName)

	process.entrypoint = getEntrypointPath(ctx.Running.AppDir, ctx.Running.Entrypoint)
	process.runDir = ctx.Running.RunDir
	process.pidFile = project.GetInstancePidFile(ctx, instanceName)
	process.workDir = project.GetInstanceWorkDir(ctx, instanceName)
	process.logDir = ctx.Running.LogDir
	process.logFile = project.GetInstanceLogFile(ctx, instanceName)
	process.consoleSock = project.GetInstanceConsoleSock(ctx, instanceName)

	process.notifySockPath = project.GetInstanceNotifySockPath(ctx, instanceName)

	process.env = append(process.env,
		formatEnv("TARANTOOL_APP_NAME", ctx.Project.Name),
		formatEnv("TARANTOOL_INSTANCE_NAME", instanceName),
		formatEnv("TARANTOOL_CFG", ctx.Running.ConfPath),
		formatEnv("TARANTOOL_CONSOLE_SOCK", process.consoleSock),
		formatEnv("TARANTOOL_PID_FILE", process.pidFile),
		formatEnv("TARANTOOL_WORKDIR", process.workDir),
	)

	process.SetPidAndStatus()

	return &process
}

func NewStateboardProcess(ctx *context.Ctx) *Process {
	var process Process

	process.ID = ctx.Project.StateboardName

	process.entrypoint = getEntrypointPath(ctx.Running.AppDir, ctx.Running.StateboardEntrypoint)
	process.runDir = ctx.Running.RunDir
	process.pidFile = project.GetStateboardPidFile(ctx)
	process.workDir = project.GetStateboardWorkDir(ctx)
	process.logDir = ctx.Running.LogDir
	process.logFile = project.GetStateboardLogFile(ctx)
	process.consoleSock = project.GetStateboardConsoleSock(ctx)

	process.notifySockPath = project.GetStateboardNotifySockPath(ctx)

	process.env = append(process.env,
		formatEnv("TARANTOOL_APP_NAME", ctx.Project.StateboardName),
		formatEnv("TARANTOOL_CFG", ctx.Running.ConfPath),
		formatEnv("TARANTOOL_CONSOLE_SOCK", process.consoleSock),
		formatEnv("TARANTOOL_PID_FILE", process.pidFile),
		formatEnv("TARANTOOL_WORKDIR", process.workDir),
	)

	process.SetPidAndStatus()

	return &process
}

func (process *Process) Log(follow bool, n int) error {
	if _, err := os.Stat(process.logFile); err != nil {
		return fmt.Errorf("Failed to use process log file: %s", err)
	}

	offset, err := common.GetLastNLinesBegin(process.logFile, n)
	if err != nil {
		return fmt.Errorf("Failed to find offset in file: %s", err)
	}

	t, err := tail.TailFile(process.logFile, tail.Config{
		Follow:    follow,
		MustExist: true,
		Location: &tail.SeekInfo{
			Offset: offset,
			Whence: io.SeekStart,
		},
		Logger: tail.DiscardingLogger,
	})
	if err != nil {
		return fmt.Errorf("Failed to get logs tail: %s", err)
	}

	writer, err := newColorizedWriter(process)
	if err != nil {
		return fmt.Errorf("Failed to create colorized logs writer: %s", err)
	}

	for line := range t.Lines {
		if _, err := writer.Write([]byte(line.Text + "\n")); err != nil {
			return fmt.Errorf("Failed to write log line: %s", err)
		}
	}

	return nil
}

func (process *Process) Clean() ProcessRes {
	res := ProcessRes{
		ProcessID: process.ID,
	}

	pathsToDelete := []string{
		process.logFile,
		process.workDir,
		process.consoleSock,
		process.notifySockPath,
		// PID file isn't deleted
		// since it's used by Cartridge CLI to start and stop instance
	}

	var nonExistedFiles []string
	var errors []string
	var skipped = true

	for _, path := range pathsToDelete {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			nonExistedFiles = append(nonExistedFiles, path)
			continue
		} else if err != nil {
			errors = append(errors, err.Error())
			continue
		} else if err := os.RemoveAll(path); err != nil {
			errors = append(errors, err.Error())
			continue
		}

		// skipped result is returned only if all files exists
		skipped = false
	}

	if len(errors) > 0 {
		res.Res = procResFailed
		res.Error = fmt.Errorf("Failed to remove some files: %s", strings.Join(errors, ", "))
	} else if skipped {
		res.Res = procResSkipped
	} else {
		res.Res = procResOk
	}

	if len(nonExistedFiles) > 0 {
		verb := "don't"
		if len(nonExistedFiles) == 1 {
			verb = "doesn't"
		}

		err := fmt.Errorf("%s %s exist", strings.Join(nonExistedFiles, ", "), verb)
		if res.Error != nil {
			res.Error = fmt.Errorf("%s. %s", res.Error, err)
		} else {
			res.Error = err
		}
	}

	return res
}
