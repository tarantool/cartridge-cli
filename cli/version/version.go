package version

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	goVersion "github.com/hashicorp/go-version"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/templates"
)

var (
	gitTag       string
	gitCommit    string
	versionLabel string
)

const (
	unknownVersion           = "<unknown>"
	cliVersionTitle          = "Tarantool Cartridge CLI"
	cartridgeVersionTitle    = "Tarantool Cartridge"
	cartridgeVersionGetError = "Failed to get the version of the Cartridge"
)

func formatVersion(template string, templateArgs map[string]string) string {
	versionMsg, err := templates.GetTemplatedStr(&template, templateArgs)

	if err != nil {
		panic(err)
	}

	return versionMsg
}

func getRocksVersions(projectPath string) (map[string]string, error) {
	var rocksVersionsMap map[string]string
	var err error

	if fileInfo, err := os.Stat(projectPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s. Your project path is invalid", cartridgeVersionGetError)
	} else if err != nil {
		return nil, fmt.Errorf("%s. Failed to use specified project path: %s", cartridgeVersionGetError, err)
	} else if !fileInfo.IsDir() {
		return nil, fmt.Errorf("%s. %s is not a directory", cartridgeVersionGetError, projectPath)
	}

	if rockspecPath, err := common.FindRockspec(projectPath); err != nil {
		return nil, fmt.Errorf("%s. %s", cartridgeVersionGetError, err)
	} else if rockspecPath == "" {
		return nil, fmt.Errorf("%s. Project path %s is not a project", cartridgeVersionGetError, projectPath)
	}

	if rocksVersionsMap, err = common.LuaGetRocksVersions(projectPath); err != nil {
		return nil, fmt.Errorf("%s. %s", cartridgeVersionGetError, err)
	}

	if len(rocksVersionsMap) == 0 {
		return nil, fmt.Errorf(`%s. Looks like your project directory
does not contain a .rocks directory... Did you built your project?`, cartridgeVersionGetError)
	}

	if rocksVersionsMap["cartridge"] == "" {
		return rocksVersionsMap, fmt.Errorf("%s. Are dependencies in .rocks directory correct?", cartridgeVersionGetError)
	}

	return rocksVersionsMap, nil
}

func buildCartridgeVersionString(rocksVersions map[string]string) string {
	version := rocksVersions["cartridge"]
	if version == "" {
		version = unknownVersion
	}

	return formatVersion(cartridgeVersionTmpl, map[string]string{
		"Title":   cartridgeVersionTitle,
		"Version": version,
	})
}

func buildRocksVersionString(rocksVersions map[string]string) string {
	var versionParts []string
	versionParts = append(versionParts, "\nRocks")

	for rock, version := range rocksVersions {
		// We have to skip cartridge rock - we print info about
		// this rock in function above.
		if rock != "cartridge" {
			versionParts = append(versionParts, fmt.Sprintf("%s %s", rock, version))
		}
	}

	return strings.Join(versionParts, "\n ")
}

func BuildCliVersionString() string {
	var version string

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

	return formatVersion(cliVersionTmpl, map[string]string{
		"Title":   cliVersionTitle,
		"Version": version,
		"OS":      runtime.GOOS,
		"Arch":    runtime.GOARCH,
		"Commit":  gitCommit,
	})
}

func BuildVersionString(projectPath string, needRocks bool) (string, error) {
	var versionParts []string
	var rocksVersions map[string]string
	var err error

	versionParts = append(versionParts, BuildCliVersionString())
	rocksVersions, err = getRocksVersions(projectPath)
	// If we get error, we anyway have to print <unknow>
	// version of Cartridge. And only after this, we return from this function.
	versionParts = append(versionParts, buildCartridgeVersionString(rocksVersions))

	if err != nil && needRocks {
		return strings.Join(versionParts, "\n"), err
	}

	if needRocks {
		versionParts = append(versionParts, buildRocksVersionString(rocksVersions))
	}

	return strings.Join(versionParts, "\n"), err
}

var (
	cliVersionTmpl = `{{ .Title }}
 Version:	{{ .Version }}
 OS/Arch: 	{{ .OS }}/{{ .Arch }}
 Git commit:	{{ .Commit }}
`

	cartridgeVersionTmpl = `{{ .Title }}
 Version:	{{ .Version }}`
)
