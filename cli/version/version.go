package version

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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
)

func buildRocksVersionString(projectPath string) string {
	var versionParts []string
	var rocksVersionsMap map[string]string
	var err error

	if _, err := os.Stat(filepath.Join(projectPath)); os.IsNotExist(err) {
		versionParts = append(versionParts, fmt.Sprintf("Your project path is invalid"))
		return strings.Join(versionParts, "\n ")
	}

	if rocksVersionsMap, err = common.LuaGetRocksVersions(projectPath); err != nil {
		versionParts = append(versionParts, fmt.Sprintf("%s. See --project-path flag", err))
		return strings.Join(versionParts, "\n ")
	}

	if len(rocksVersionsMap) == 0 {
		versionParts = append(versionParts, fmt.Sprintf(`Looks like your project directory does not contain a .rocks directory...
See --project-path flag`))
		return strings.Join(versionParts, "\n ")
	}

	for k, v := range rocksVersionsMap {
		versionParts = append(versionParts, fmt.Sprintf("%s v%s", k, v))
	}

	return strings.Join(versionParts, "\n")
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
	versionParts = append(versionParts, BuildCliVersionString())
	if needRocks {
		versionParts = append(versionParts, buildRocksVersionString(projectPath))
	}

	return strings.Join(versionParts, "\n\n")
}
