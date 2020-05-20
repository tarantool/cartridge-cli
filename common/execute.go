package common

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/briandowns/spinner"
)

var (
	spinnerPicture    = spinner.CharSets[9]
	spinnerUpdateTime = 100 * time.Millisecond
)

const (
	ready = 1
)

// startAndWaitCommand executes command
// and sends `ready` flag to the channel before return
func startAndWaitCommand(cmd *exec.Cmd, c chan int, wg *sync.WaitGroup, err *error) {
	defer wg.Done()
	defer func() { c <- ready }() // say that command is complete

	if *err = cmd.Start(); *err != nil {
		return
	}

	if *err = cmd.Wait(); *err != nil {
		return
	}
}

// startCommandSpinner starts running spinner
// until `ready` flag is received from the channel
func startCommandSpinner(c chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	s := spinner.New(spinnerPicture, spinnerUpdateTime)
	s.Start()

	// wait for the command to complete
	_ = <-c

	s.Stop()
}

// RunCommand runs specified command and returns an error
// If showOutput is set to true, command output is shown
// Else spinner is shown while command is running
func RunCommand(cmd *exec.Cmd, dir string, showOutput bool) error {
	var err error
	var wg sync.WaitGroup
	c := make(chan int, 1)

	cmd.Dir = dir
	if showOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		wg.Add(1)
		go startCommandSpinner(c, &wg)
	}

	wg.Add(1)
	go startAndWaitCommand(cmd, c, &wg, &err)

	wg.Wait()

	return err
}

// RunHook runs specified hook and returns an error
// If showOutput is set to true, command output is shown
func RunHook(hookPath string, showOutput bool) error {
	hookName := filepath.Base(hookPath)
	hookDir := filepath.Dir(hookPath)

	if isExec, err := IsExecOwner(hookPath); err != nil {
		return fmt.Errorf("Failed go check hook file `%s`: %s", hookName, err)
	} else if !isExec {
		return fmt.Errorf("Hook `%s` should be executable", hookName)
	}

	hookCmd := exec.Command(hookPath)
	err := RunCommand(hookCmd, hookDir, showOutput)
	if err != nil {
		return fmt.Errorf("Failed to run hook `%s`: %s", hookName, err)
	}

	return nil
}

// GetOutput runs specified command and returns it's stdout
func GetOutput(cmd *exec.Cmd, dir *string) (string, error) {
	var stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if dir != nil {
		cmd.Dir = *dir
	}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf(
			"Failed to run \n%s\n\n Stderr: %s\n\n Stdout: %s",
			cmd.String(), stderrBuf.String(), stdoutBuf.String(),
		)
	}

	return stdoutBuf.String(), nil
}
