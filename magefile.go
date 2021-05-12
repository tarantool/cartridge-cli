// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// can be overwritten by GOEXE
var goExe = "go"

// can be overwritten by PY3EXE
var py3Exe = "python3"

// can be overwritten by CLIEXE
var cliExe = "cartridge"

var goPackageName = "github.com/tarantool/cartridge-cli/cli"
var packagePath = "./cli"

var generateModePath = filepath.Join(packagePath, "codegen", "generate_code.go")
var generatedFilesPath = fmt.Sprintf("./%s", filepath.Join(packagePath, "codegen", "static"))

var generatedFSFile = "cartridge_vfsdata_gen.go"
var generatedModeFile = "create_cartrdige_template_filemodes_gen.go"

var completionPath = "./completion"

var tmpPath = "./tmp"
var sdkDirName = "tarantool-enterprise"
var sdkDirPath = filepath.Join(tmpPath, sdkDirName)

func getBuildEnv() map[string]string {
	var err error

	var curDir string
	var gitTag string
	var gitCommit string

	if curDir, err = os.Getwd(); err != nil {
		fmt.Printf("Failed to get current directory: %s\n", err)
	}

	if _, err := exec.LookPath("git"); err == nil {
		gitTag, _ = sh.Output("git", "describe", "--tags")
		gitCommit, _ = sh.Output("git", "rev-parse", "--short", "HEAD")

	}

	versionLabel := os.Getenv("VERSION_LABEL")

	return map[string]string{
		"PACKAGE":       goPackageName,
		"GIT_TAG":       gitTag,
		"GIT_COMMIT":    gitCommit,
		"VERSION_LABEL": versionLabel,
		"PWD":           curDir,
	}
}

var ldflags = []string{
	"-s", "-w",
	"-X ${PACKAGE}/version.gitTag=${GIT_TAG}",
	"-X ${PACKAGE}/version.gitCommit=${GIT_COMMIT}",
	"-X ${PACKAGE}/version.versionLabel=${VERSION_LABEL}",
}
var ldflagsStr = strings.Join(ldflags, " ")

var asmflags = "all=-trimpath=${PWD}"
var gcflags = "all=-trimpath=${PWD}"

func init() {
	var err error

	if specifiedGoExe := os.Getenv("GOEXE"); specifiedGoExe != "" {
		goExe = specifiedGoExe
	}

	if specifiedCliExe := os.Getenv("CLIEXE"); specifiedCliExe != "" {
		cliExe = specifiedCliExe
	} else {
		if cliExe, err = filepath.Abs(cliExe); err != nil {
			panic(err)
		}
	}

	// We want to use Go 1.11 modules even if the source lives inside GOPATH.
	// The default is "auto".
	os.Setenv("GO111MODULE", "on")
}

// Run go vet and flake8
func Lint() error {
	fmt.Println("Generating Go code...")
	mg.Deps(GenerateGoCode)

	fmt.Println("Running go vet...")
	if err := sh.RunV(goExe, "vet", packagePath); err != nil {
		return err
	}

	fmt.Println("Running flake8...")
	if err := sh.RunV(py3Exe, "-m", "flake8", "--exclude", os.Getenv("FLAKE_EXCLUDE")); err != nil {
		return err
	}

	fmt.Println("Running luacheck for test projects files...")
	if err := sh.RunV(".rocks/bin/luacheck", "test/files"); err != nil {
		return err
	}

	return nil
}

// Run unit tests
func Unit() error {
	fmt.Println("Running unit tests...")
	mg.Deps(GenerateGoCode)

	if mg.Verbose() {
		return sh.RunV(goExe, "test", "-v", "./cli/...")
	} else {
		return sh.RunV(goExe, "test", "./cli/...")
	}
}

// Run integration tests
func Integration() error {
	fmt.Println("Running integration tests...")
	return sh.RunV(py3Exe, "-m", "pytest", "test/integration")
}

// Run examples tests
func TestExamples() error {
	fmt.Println("Running examples tests...")
	return sh.RunV(py3Exe, "-m", "pytest", "test/examples")
}

// Run e2e tests
func E2e() error {
	fmt.Println("Running e2e tests...")
	return sh.RunV(py3Exe, "-m", "pytest", "test/e2e")
}

// Run all tests
func Test() {
	mg.SerialDeps(Lint, Unit, Integration, TestExamples, E2e)
}

// Build cartridge-cli executable
func Build() error {
	var err error

	fmt.Println("Building...")

	mg.Deps(GenerateGoCode)

	err = sh.RunWith(
		getBuildEnv(), goExe, "build",
		"-o", cliExe,
		"-ldflags", ldflagsStr,
		"-asmflags", asmflags,
		"-gcflags", gcflags,
		packagePath,
	)

	if err != nil {
		return fmt.Errorf("Failed to build cartridge-cli executable: %s", err)
	}

	return nil
}

// Generate Go code that statically implements filesystem
// and map with modes for that filesystem.
func GenerateGoCode() error {
	err := sh.RunWith(
		getBuildEnv(), goExe,
		"generate", "-tags=dev",
		generatedFilesPath,
	)

	if err != nil {
		return err
	}

	err = sh.RunWith(
		getBuildEnv(), goExe,
		"run", generateModePath,
	)

	if err != nil {
		return err
	}

	return nil
}

// Generate completion scripts for bash and zsh
func GenCompletion() error {
	if err := Build(); err != nil {
		return err
	}

	fmt.Println("Generate autocompletion...")

	if err := sh.Run(cliExe, "gen", "completion"); err != nil {
		return fmt.Errorf("Failed to generate autocompletion scripts: %s", err)
	}

	return nil
}

// Download Tarantool Enterprise to tmp/tarantool-enterprise dir
func Sdk() error {
	if _, err := os.Stat(sdkDirPath); os.IsNotExist(err) {
		if err := downloadSdk(); err != nil {
			return err
		}
	} else if err != nil {
		return fmt.Errorf("Failed to check if SDK exists: %s", err)
	} else {
		fmt.Printf("Found Tarantool Enterprise SDK: %s\n", sdkDirPath)
	}

	fmt.Printf("Run `source %s/env.sh` to activate Tarantool Enterprise\n", sdkDirPath)

	return nil
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")

	os.Remove(filepath.Join(generatedFilesPath, generatedFSFile))
	os.Remove(filepath.Join(generatedFilesPath, generatedModeFile))

	os.RemoveAll(cliExe)
	os.RemoveAll(completionPath)
}

func downloadSdk() error {
	bundleVersion := os.Getenv("BUNDLE_VERSION")
	if bundleVersion == "" {
		return fmt.Errorf("Please, specify BUNDLE_VERSION")
	}

	downloadToken := os.Getenv("DOWNLOAD_TOKEN")
	if downloadToken == "" {
		return fmt.Errorf("Please, specify DOWNLOAD_TOKEN")
	}

	archivedSDKName := fmt.Sprintf("tarantool-enterprise-bundle-%s.tar.gz", bundleVersion)
	sdkDownloadUrl := fmt.Sprintf(
		"https://tarantool:%s@download.tarantool.io/enterprise/%s",
		downloadToken,
		archivedSDKName,
	)

	fmt.Printf("Download Tarantool Enterprise SDK %s...\n", bundleVersion)

	archivedSDKPath := filepath.Join(tmpPath, archivedSDKName)
	if err := downloadFile(sdkDownloadUrl, archivedSDKPath); err != nil {
		return fmt.Errorf("Failed to download archived SDK: %s", err)
	}
	defer os.RemoveAll(archivedSDKPath)

	fmt.Println("Unarchive Tarantool Enterprise SDK...")

	if err := sh.RunV("tar", "-xzf", archivedSDKPath, "-C", tmpPath); err != nil {
		return fmt.Errorf("Failed to unarchive SDK: %s")
	}

	return nil
}
