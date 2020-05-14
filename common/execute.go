package common

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RunCommand runs specified command and returns an error
// If showOutput is set to true, command output is shown
func RunCommand(cmd *exec.Cmd, dir string, showOutput bool) error {
	cmd.Dir = dir

	if showOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
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
