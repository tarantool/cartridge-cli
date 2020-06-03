package running

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/cli/project"
)

type Process struct {
	ID string

	entrypoint string
	runDir     string
	pidFile    string
	env        []string

	pid       int
	osProcess *os.Process
	writer    io.Writer
}

type ProcessesSet map[string]*Process

func (set *ProcessesSet) Add(processes ...*Process) error {
	for _, process := range processes {
		if _, found := (*set)[process.ID]; found {
			return fmt.Errorf("Duplicate process ID: %s", process.ID)
		}

		(*set)[process.ID] = process
	}

	return nil
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

	// create run dir
	if err := os.MkdirAll(process.runDir, 0755); err != nil {
		return fmt.Errorf("Failed to initialize run dir: %s", err)
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

	process.entrypoint = projectCtx.Entrypoint
	process.runDir = projectCtx.RunDir

	process.pidFile = project.GetInstancePidFile(projectCtx, instanceName)
	consoleSock := project.GetInstanceConsoleSock(projectCtx, instanceName)

	process.env = append(process.env,
		formatEnv("TARANTOOL_APP_NAME", projectCtx.Name),
		formatEnv("TARANTOOL_INSTANCE_NAME", instanceName),
		formatEnv("TARANTOOL_CFG", projectCtx.ConfDir), // XXX: rename to ConfPath
		formatEnv("TARANTOOL_CONSOLE_SOCK", consoleSock),
	)

	process.writer = newProcessWriter(&process)

	return &process
}

func NewStateboardProcess(projectCtx *project.ProjectCtx) *Process {
	var process Process

	process.ID = projectCtx.StateboardName

	process.entrypoint = projectCtx.StateboardEntrypoint
	process.runDir = projectCtx.RunDir

	process.pidFile = project.GetStateboardPidFile(projectCtx)
	consoleSock := project.GetStateboardConsoleSock(projectCtx)

	process.env = append(process.env,
		formatEnv("TARANTOOL_APP_NAME", projectCtx.StateboardName),
		formatEnv("TARANTOOL_CFG", projectCtx.ConfDir), // XXX: rename to ConfPath
		formatEnv("TARANTOOL_CONSOLE_SOCK", consoleSock),
	)

	process.writer = newProcessWriter(&process)

	return &process
}

func formatEnv(key, value string) string {
	return fmt.Sprintf("%s=%s", key, value)
}
