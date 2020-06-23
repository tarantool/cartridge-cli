package main

import (
	"github.com/apex/log"

	"github.com/tarantool/cartridge-cli/cli/commands"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatalf("%s", project.InternalError("Unhandled internal error: %s", r))
		}
	}()

	commands.Execute()
}
