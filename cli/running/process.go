package running

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/cli/project"
)

type Process struct {
	ID string

	entrypoint string
	runDir     string
	workDir    string
	pidFile    string
	env        []string

	pid       int
	osProcess *os.Process
	writer    io.Writer
}

type ProcessesSet []*Process

func (set *ProcessesSet) Add(processes ...*Process) {
	*set = append(*set, processes...)
}

func (set *ProcessesSet) Start(daemonize bool) error {
	errCh := make(chan error)

	if daemonize {
		panic("Staring in detached mode is not supported yet")
	}

	for _, process := range *set {
		log.Infof("Starting %s", process.ID)

		go func(process *Process) {
			if err := process.StartInteractive(); err != nil {
				errCh <- fmt.Errorf("%s exited: %s", process.ID, err)
			}
		}(process)

		time.Sleep(1 * time.Second)
	}

	for i := 0; i < len(*set); i++ {
		select {
		case err := <-errCh:
			log.Errorf(err.Error())
		}
	}

	return fmt.Errorf("All instances exited")
}

func (process *Process) StartInteractive() error {
	var cmd *exec.Cmd

	ctx := context.Background()
	cmd = exec.CommandContext(ctx, "tarantool", process.entrypoint)

	cmd.Env = append(os.Environ(), process.env...)
	cmd.Dir = process.workDir

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

	cmd.Stdout = process.writer
	cmd.Stderr = process.writer

	if err := cmd.Run(); err != nil {
		return err
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
	consoleSock := project.GetInstanceConsoleSock(projectCtx, instanceName)

	process.env = append(process.env,
		formatEnv("TARANTOOL_APP_NAME", projectCtx.Name),
		formatEnv("TARANTOOL_INSTANCE_NAME", instanceName),
		formatEnv("TARANTOOL_CFG", projectCtx.ConfPath),
		formatEnv("TARANTOOL_CONSOLE_SOCK", consoleSock),
		formatEnv("TARANTOOL_PID_FILE", process.pidFile),
		formatEnv("TARANTOOL_WORKDIR", process.workDir),
	)

	process.writer = newProcessWriter(&process)

	return &process
}

func NewStateboardProcess(projectCtx *project.ProjectCtx) *Process {
	var process Process

	process.ID = projectCtx.StateboardName

	process.entrypoint = filepath.Join(projectCtx.Path, projectCtx.StateboardEntrypoint)
	process.runDir = projectCtx.RunDir
	process.pidFile = project.GetStateboardPidFile(projectCtx)
	process.workDir = project.GetStateboardWorkDir(projectCtx)
	consoleSock := project.GetStateboardConsoleSock(projectCtx)

	process.env = append(process.env,
		formatEnv("TARANTOOL_APP_NAME", projectCtx.StateboardName),
		formatEnv("TARANTOOL_CFG", projectCtx.ConfPath),
		formatEnv("TARANTOOL_CONSOLE_SOCK", consoleSock),
		formatEnv("TARANTOOL_PID_FILE", process.pidFile),
		formatEnv("TARANTOOL_WORKDIR", process.workDir),
	)

	process.writer = newProcessWriter(&process)

	return &process
}

func formatEnv(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}
