package version

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/apex/log"
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
	cartridgeVersionGetError = "Failed to show Cartridge version"
	rocksVersionsGetError    = "Failed to show Cartridge and other rocks versions"
)

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

func PrintVersionString(projectPath string, projectPathIsSet bool, showRocksVersions bool) error {
	fmt.Println(BuildCliVersionString())

	if err := printCartridgeVersion(projectPath); err != nil {
		currentErrorString := cartridgeVersionGetError
		if showRocksVersions {
			currentErrorString = rocksVersionsGetError
		}

		if projectPathIsSet {
			return fmt.Errorf("%s: %s", currentErrorString, err)
		}

		log.Warnf("%s: %s. See --project-path flag, to specify path to project", currentErrorString, err)
		return nil
	}

	if showRocksVersions {
		if err := printRocksVersion(projectPath); err != nil {
			return fmt.Errorf("%s: %s", rocksVersionsGetError, err)
		}
	}

	return nil
}

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
		return nil, fmt.Errorf("Specified project path doesn't exist")
	} else if err != nil {
		return nil, fmt.Errorf("Impossible to use specified project path: %s", err)
	} else if !fileInfo.IsDir() {
		return nil, fmt.Errorf("Specified project path %s is not a directory", projectPath)
	}

	if rocksVersionsMap, err = common.LuaGetRocksVersions(projectPath); err != nil {
		return nil, err
	}

	return rocksVersionsMap, nil
}

func printCartridgeVersion(projectPath string) error {
	var rocksVersions map[string]string
	var err error

	if rocksVersions, err = getRocksVersions(projectPath); err != nil {
		return err
	}

	if rockspecPath, err := common.FindRockspec(projectPath); err != nil {
		return err
	} else if rockspecPath == "" {
		return fmt.Errorf("Project path %s is not a project", projectPath)
	}

	version := rocksVersions["cartridge"]
	if version == "" {
		return fmt.Errorf("Are dependencies in .rocks directory correct?")
	}

	cartridgeVersion := formatVersion(cartridgeVersionTmpl, map[string]string{
		"Title":   cartridgeVersionTitle,
		"Version": version,
	})

	fmt.Println(cartridgeVersion)
	return err
}

func printRocksVersion(projectPath string) error {
	var versionParts []string
	var rocksVersions map[string]string
	var err error

	if rocksVersions, err = getRocksVersions(projectPath); err != nil {
		return err
	}

	if len(rocksVersions) == 0 {
		return fmt.Errorf(`Looks like your project directory
does not contain a .rocks directory... Did you built your project?`)
	}

	versionParts = append(versionParts, "\nRocks")

	for rock, version := range rocksVersions {
		// We have to skip cartridge rock - we print info about
		// this rock in function above.
		if rock != "cartridge" {
			versionParts = append(versionParts, fmt.Sprintf("%s %s", rock, version))
		}
	}

	fmt.Println(strings.Join(versionParts, "\n "))
	return nil
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
