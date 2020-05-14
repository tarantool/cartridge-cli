package common

import (
	"fmt"
	"os"
	"os/exec"
)

// Prompt a value with given text and default value
func Prompt(text, defaultValue string) string {
	if defaultValue == "" {
		fmt.Printf("%s: ", text)
	} else {
		fmt.Printf("%s [%s]: ", text, defaultValue)
	}

	var value string
	fmt.Scanf("%s", &value)

	if value == "" {
		value = defaultValue
	}

	return value
}

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
