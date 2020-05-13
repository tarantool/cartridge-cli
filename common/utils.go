package common

import (
	"fmt"
	"os"
	"os/exec"
)

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

func RunCommand(cmd *exec.Cmd, dir string, showOutput bool) error {
	cmd.Dir = dir

	if showOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}
