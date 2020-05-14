package pack

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/project"
)

func PackProject(projectCtx project.ProjectCtx) error {
	log.Infof("Packing %s into %s", projectCtx.Name, projectCtx.PackType)

	// projectCtx.BuildDir = projectCtx.Path

	if projectCtx.BuildInDocker {
		log.Fatal("Not implemented yet")
	}

	fmt.Printf("%#v\n", projectCtx)

	log.Infof("Application succeessfully packed")

	return nil
}
