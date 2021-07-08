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
	var rocksVersions common.RocksVersions
	var err error

	fmt.Println(BuildCliVersionString())

	currentErrorString := cartridgeVersionGetError
	if showRocksVersions {
		currentErrorString = rocksVersionsGetError
	}

	if rocksVersions, err = LuaGetRocksVersions(projectPath); err != nil {
		return fmt.Errorf("%s: %s", currentErrorString, err)
	}

	if err := printCartridgeVersion(projectPath, rocksVersions, showRocksVersions); err != nil {
		if projectPathIsSet {
			return fmt.Errorf("%s: %s", currentErrorString, err)
		}

		log.Warnf("%s: %s. See --project-path flag, to specify path to project", currentErrorString, err)
		return nil
	}

	if showRocksVersions {
		if err := printRocksVersion(projectPath, rocksVersions); err != nil {
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

func LuaGetRocksVersions(projectPath string) (common.RocksVersions, error) {
	var rocksVersions common.RocksVersions
	var err error

	if fileInfo, err := os.Stat(projectPath); os.IsNotExist(err) {
		return rocksVersions, fmt.Errorf("Specified project path doesn't exist")
	} else if err != nil {
		return rocksVersions, fmt.Errorf("Impossible to use specified project path: %s", err)
	} else if !fileInfo.IsDir() {
		return rocksVersions, fmt.Errorf("Specified project path %s is not a directory", projectPath)
	}

	if rocksVersions, err = common.LuaGetRocksVersions(projectPath); err != nil {
		return rocksVersions, err
	}

	return rocksVersions, nil
}

func printCartridgeVersion(projectPath string, rocksVersions common.RocksVersions, showRocksVersions bool) error {
	if rockspecPath, err := common.FindRockspec(projectPath); err != nil {
		return err
	} else if rockspecPath == "" {
		return fmt.Errorf("Project path %s is not a project", projectPath)
	}

	versions := rocksVersions["cartridge"]
	if versions == nil {
		return fmt.Errorf("Are dependencies in .rocks directory correct?")
	}

	cartridgeVersion := formatVersion(cartridgeVersionTmpl, map[string]string{
		"Title":   cartridgeVersionTitle,
		"Version": strings.Join(versions, ", "),
	})

	fmt.Print(cartridgeVersion)
	if !showRocksVersions && len(versions) > 1 {
		fmt.Println()
		log.Warnf("Found multiple versions of Cartridge in rocks manifest")
	}

	return nil
}

func printRocksVersion(projectPath string, rocksVersions common.RocksVersions) error {
	var versionParts []string
	duplicatesFound := false

	if len(rocksVersions) == 0 {
		return fmt.Errorf(`Looks like your project directory
does not contain a .rocks directory... Did you built your project?`)
	}

	versionParts = append(versionParts, "\nRocks")

	for rock, versions := range rocksVersions {
		// We have to skip cartridge rock - we print info about
		// this rock in function above.
		if rock != "cartridge" {
			versionParts = append(versionParts, fmt.Sprintf("%s %s", rock, strings.Join(versions, ", ")))
			if len(versions) > 1 {
				duplicatesFound = true
			}
		}
	}

	fmt.Println(strings.Join(versionParts, "\n "))
	if duplicatesFound {
		fmt.Println()
		log.Warnf("Found multiple versions in rocks manifest")
	}

	return nil
}

var (
	cliVersionTmpl = `{{ .Title }}
 Version:	{{ .Version }}
 OS/Arch: 	{{ .OS }}/{{ .Arch }}
 Git commit:	{{ .Commit }}
`

	cartridgeVersionTmpl = `{{ .Title }}
 Version:	{{ .Version }}
`
)
