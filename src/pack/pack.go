package pack

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/tarantool/cartridge-cli/src/common"
	"github.com/tarantool/cartridge-cli/src/project"
)

var (
	packers = map[string]func(*project.ProjectCtx) error{
		tgzType: packTgz,
		debType: packDeb,
		rpmType: packRpm,
	}
)

const (
	tgzType = "tgz"
	rpmType = "rpm"
	debType = "deb"

	defaultBuildDockerfile = "Dockerfile.build.cartridge"
)

// Run packs application into project.PackType distributable
func Run(projectCtx *project.ProjectCtx) error {
	// check context
	if err := checkCtx(projectCtx); err != nil {
		// TODO: format internal error
		panic(err)
	}

	// set build base Dockerfile
	if projectCtx.BuildInDocker {
		if projectCtx.BuildFrom == "" {
			defaultBuildDockerfilePath := filepath.Join(projectCtx.Path, defaultBuildDockerfile)
			if _, err := os.Stat(defaultBuildDockerfilePath); err == nil {
				log.Debugf("Default build Dockerfile is used: %s", defaultBuildDockerfilePath)

				projectCtx.BuildFrom = defaultBuildDockerfilePath
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("Failed to use default build Dockerfile: %s", err)
			}
		}
	}

	// get packer function
	packer, found := packers[projectCtx.PackType]
	if !found {
		return fmt.Errorf("Unsupported distribution type: %s", projectCtx.PackType)
	}

	if _, err := os.Stat(projectCtx.Path); err != nil {
		return fmt.Errorf("Failed to use path %s: %s", projectCtx.Path, err)
	}

	checkPackRecommendedBinaries()

	projectCtx.PackID = common.RandomString(10)
	projectCtx.BuildID = projectCtx.PackID

	// get and normalize version
	if err := detectVersion(projectCtx); err != nil {
		return err
	}

	// check if app has stateboard entrypoint
	stateboardEntrypointPath := filepath.Join(projectCtx.Path, project.StateboardEntrypointName)
	if _, err := os.Stat(stateboardEntrypointPath); err == nil {
		projectCtx.WithStateboard = true
	} else if os.IsNotExist(err) {
		projectCtx.WithStateboard = false
	} else {
		return fmt.Errorf("Failed to get stateboard entrypoint stat: %s", err)
	}

	// set result package path
	curDir, err := os.Getwd()
	if err != nil {
		return err
	}
	projectCtx.ResPackagePath = filepath.Join(curDir, getPackageFullname(projectCtx))

	// tmp directory
	if err := detectTmpDir(projectCtx); err != nil {
		return err
	}

	log.Infof("Temporary directory is set to %s\n", projectCtx.TmpDir)

	if err := initTmpDir(projectCtx); err != nil {
		return err
	}

	defer project.RemoveTmpPath(projectCtx.TmpDir, projectCtx.Debug)

	// call packer
	log.Infof("Packing %s into %s", projectCtx.Name, projectCtx.PackType)

	if err := packer(projectCtx); err != nil {
		return err
	}

	log.Infof("Application succeessfully packed")

	return nil
}

func checkCtx(projectCtx *project.ProjectCtx) error {
	if projectCtx.Path == "" {
		return fmt.Errorf("Missed project path")
	}

	if projectCtx.TarantoolDir == "" {
		return fmt.Errorf("Missed Tarantool directory path")
	}

	if projectCtx.TarantoolVersion == "" {
		return fmt.Errorf("Missed Tarantool version")
	}

	if projectCtx.PackType == "" {
		return fmt.Errorf("Missed distribution type")
	}

	return nil
}

func checkPackRecommendedBinaries() {
	var recommendedBinaries = []string{
		"git",
	}

	// check recommended binaries
	for _, binary := range recommendedBinaries {
		if _, err := exec.LookPath(binary); err != nil {
			log.Warnf("%s binary is recommended to pack application", binary)
		}
	}
}
