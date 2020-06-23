package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/cli/commands"
	"github.com/tarantool/cartridge-cli/cli/project"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(project.InternalError("Unhandled internal error: %s", r))
		}
	}()

	commands.Execute()
}
