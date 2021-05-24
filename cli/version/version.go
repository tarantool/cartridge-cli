package version

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/apex/log"
	goVersion "github.com/hashicorp/go-version"

	"github.com/tarantool/cartridge-cli/cli/common"
)

var (
	gitTag       string
	gitCommit    string
	versionLabel string
)

const (
	unknownVersion = "<unknown>"
	cliName        = "Tarantool Cartridge CLI"
	cartridgeName  = "Tarantool Cartridge"
	errorStr       = "Failed to get the version of the Cartridge"
)

func getRocksVersions(projectPath string) (map[string]string, error) {
	var rocksVersionsMap map[string]string
	var err error

	if _, err := os.Stat(filepath.Join(projectPath)); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s. Your project path is invalid", errorStr)
	}

	if rocksVersionsMap, err = common.LuaGetRocksVersions(projectPath); err != nil {
		return nil, fmt.Errorf("%s. %s. See --project-path flag", err, errorStr)
	}

	if len(rocksVersionsMap) == 0 {
		return nil, fmt.Errorf(`%s. Looks like your project directory
does not contain a .rocks directory... See --project-path flag`, errorStr)
	}

	if rocksVersionsMap["cartridge"] == "" {
		return nil, fmt.Errorf("%s. Are dependencies in .rocks directory correct?", errorStr)
	}

	return rocksVersionsMap, nil
}

func buildCartridgeVersionString(rocksVersions map[string]string) string {
	var versionParts []string
	versionParts = append(versionParts, cartridgeName)

	version := rocksVersions["cartridge"]
	if version == "" {
		versionParts = append(versionParts, fmt.Sprintf("Version:\t%s", unknownVersion))
	} else {
		versionParts = append(versionParts, fmt.Sprintf("Version:\t%s", version))
	}

	return strings.Join(versionParts, "\n ")
}

func buildRocksVersionString(rocksVersions map[string]string) string {
	var versionParts []string
	versionParts = append(versionParts, "Rocks")

	for rock, version := range rocksVersions {
		// We have to skip cartridge rock - we print info about
		// this rock in function above. Also, we have to check
		// that the rock is really the rock because the manifest
		// file contains the project itself (for example: myapp - scm1)
		if rock != "cartridge" && version[0] >= '0' && version[0] <= '9' {
			versionParts = append(versionParts, fmt.Sprintf("%s v%s", rock, version))
		}
	}

	return strings.Join(versionParts, "\n ")
}

func BuildCliVersionString() string {
	var version string

	var versionParts []string
	versionParts = append(versionParts, cliName)

	if gitTag == "" {
		version = unknownVersion
	} else {
		if normalizedVersion, err := goVersion.NewVersion(gitTag); err != nil {
			version = gitTag
		} else {
			version = strings.Join(common.IntsToStrings(normalizedVersion.Segments()), ".")
		}

		if versionLabel != "" {
			version = fmt.Sprintf("%s/%s", version, versionLabel)
		}
	}

	versionStr := fmt.Sprintf("Version:\t%s", version)
	versionParts = append(versionParts, versionStr)

	osArchStr := fmt.Sprintf("OS/Arch:\t%s/%s", runtime.GOOS, runtime.GOARCH)
	versionParts = append(versionParts, osArchStr)

	if gitCommit != "" {
		gitCommitStr := fmt.Sprintf("Git commit:\t%s", gitCommit)
		versionParts = append(versionParts, gitCommitStr)
	}

	return strings.Join(versionParts, "\n ")
}

func BuildVersionString(projectPath string, needRocks bool) string {
	var versionParts []string
	var rocksVersions map[string]string
	var err error

	versionParts = append(versionParts, BuildCliVersionString())
	rocksVersions, err = getRocksVersions(projectPath)
	// If we get error, we anyway have to print <unknow>
	// version of Cartridge. And only after this, we return from this function.
	versionParts = append(versionParts, buildCartridgeVersionString(rocksVersions))

	if err != nil {
		log.Warnf("%s\n", err)
		return strings.Join(versionParts, "\n\n")
	}

	if needRocks {
		versionParts = append(versionParts, buildRocksVersionString(rocksVersions))
	}

	return strings.Join(versionParts, "\n\n")
}
