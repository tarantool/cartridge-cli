package common

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mattn/go-isatty"

	"github.com/apex/log"
	"github.com/briandowns/spinner"
)

type emptyStruct struct{}
type ReadyChan chan emptyStruct

var (
	spinnerPicture    = spinner.CharSets[9]
	spinnerUpdateTime = 100 * time.Millisecond

	ready = emptyStruct{}
)

func SendReady(c ReadyChan) {
	c <- ready
}

// startAndWaitCommand executes command
// and sends `ready` flag to the channel before return
func startAndWaitCommand(cmd *exec.Cmd, c ReadyChan, wg *sync.WaitGroup, err *error) {
	defer wg.Done()
	defer SendReady(c)

	if *err = cmd.Start(); *err != nil {
		return
	}

	if *err = cmd.Wait(); *err != nil {
		return
	}
}

// StartCommandSpinner starts running spinner
// until `ready` flag is received from the channel
func StartCommandSpinner(c ReadyChan, wg *sync.WaitGroup, prefix string) {
	defer wg.Done()

	s := spinner.New(spinnerPicture, spinnerUpdateTime)
	if prefix != "" {
		s.Prefix = fmt.Sprintf("%s ", strings.TrimSpace(prefix))
	}

	s.Start()

	// wait for the command to complete
	<-c

	s.Stop()
}

// RunCommand runs specified command and returns an error
// If showOutput is set to true, command output is shown
// Else spinner is shown while command is running
func RunCommand(cmd *exec.Cmd, dir string, showOutput bool) error {
	var err error
	var wg sync.WaitGroup
	c := make(ReadyChan, 1)

	var outputBuf *os.File

	cmd.Dir = dir
	if showOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		if outputBuf, err = ioutil.TempFile("", "out"); err != nil {
			log.Warnf("Failed to create tmp file to store command output: %s", err)
		}
		cmd.Stdout = outputBuf
		cmd.Stderr = outputBuf
		defer outputBuf.Close()
		defer os.Remove(outputBuf.Name())

		if isatty.IsTerminal(os.Stdout.Fd()) {
			wg.Add(1)
			go StartCommandSpinner(c, &wg, "")
		}
	}

	wg.Add(1)
	go startAndWaitCommand(cmd, c, &wg, &err)

	wg.Wait()

	if err != nil {
		if outputBuf != nil {
			if err := PrintFromStart(outputBuf); err != nil {
				log.Warnf("Failed to show command output: %s", err)
			}
		}

		return fmt.Errorf(
			"Failed to run \n%s\n\n%s", cmd.String(), err,
		)
	}

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
	var err error

	var stdoutBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf

	var stderrBuf *os.File
	if stderrBuf, err = ioutil.TempFile("", "err"); err != nil {
		log.Warnf("Failed to create tmp file to store command stderr: %s", err)
	} else {
		cmd.Stderr = stderrBuf
		defer stderrBuf.Close()
		defer os.Remove(stderrBuf.Name())
	}

	if dir != nil {
		cmd.Dir = *dir
	}

	if err := cmd.Run(); err != nil {
		fmt.Println("Captured stdout:")
		if _, err := io.Copy(os.Stdout, &stdoutBuf); err != nil {
			log.Warnf("Failed to show command stdout: %s", err)
		}

		if stderrBuf != nil {
			fmt.Println("Captured stderr:")
			if err := PrintFromStart(stderrBuf); err != nil {
				log.Warnf("Failed to show command stderr: %s", err)
			}
		}
		return "", fmt.Errorf(
			"Failed to run \n%s\n\n%s", cmd.String(), err,
		)
	}

	return stdoutBuf.String(), nil
}

// GetMissedBinaries returns list of binaries not found in PATH
func GetMissedBinaries(binaries ...string) []string {
	var missedBinaries []string

	for _, binary := range binaries {
		if _, err := exec.LookPath(binary); err != nil {
			missedBinaries = append(missedBinaries, binary)
		}
	}

	return missedBinaries
}

// CheckRecommendedBinaries warns if some binaries not found in PATH
func CheckRecommendedBinaries(binaries ...string) {
	missedBinaries := GetMissedBinaries(binaries...)

	if len(missedBinaries) > 0 {
		log.Warnf("Missed recommended binaries %s", strings.Join(missedBinaries, ", "))
	}
}

// CheckRequiredBinaries returns an error if some binaries not found in PATH
func CheckRequiredBinaries(binaries ...string) error {
	missedBinaries := GetMissedBinaries(binaries...)

	if len(missedBinaries) > 0 {
		return fmt.Errorf("Missed required binaries %s", strings.Join(missedBinaries, ", "))
	}

	return nil
}

// CheckTarantoolBinaries returns an error if tarantool or tarantoolctl is
// not found in PATH
func CheckTarantoolBinaries() error {
	return CheckRequiredBinaries("tarantool", "tarantoolctl")
}

// RunFunctionWithSpinner executes function and starts a spinner
// with specified prefix in a background until function returns
func RunFunctionWithSpinner(f func() error, prefix string) error {
	var err error
	var wg sync.WaitGroup
	c := make(ReadyChan, 1)

	if isatty.IsTerminal(os.Stdout.Fd()) {
		wg.Add(1)
		go StartCommandSpinner(c, &wg, prefix)
	}

	wg.Add(1)
	go func(f func() error, err *error) {
		defer wg.Done()
		defer SendReady(c)

		*err = f()
	}(f, &err)

	wg.Wait()

	return err
}
