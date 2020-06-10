// +build mage

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// can be overwritten by GOEXE
var goExe = "go"

// can be overwritten by CLIEXE
var cliExe = "cartridge"

var packageName = "github.com/tarantool/cartridge-cli/cli"
var packagePath = "./cli"

func getBuildEnv() map[string]string {
	gitTag, _ := sh.Output("git", "describe", "--tags")
	gitCommit, _ := sh.Output("git", "rev-parse", "--short", "HEAD")
	versionLabel := os.Getenv("VERSION_LABEL")

	return map[string]string{
		"PACKAGE":       packageName,
		"GIT_TAG":       gitTag,
		"GIT_COMMIT":    gitCommit,
		"VERSION_LABEL": versionLabel,
	}
}

var ldflags = []string{
	"-X ${PACKAGE}/commands.gitTag=${GIT_TAG}",
	"-X ${PACKAGE}/commands.gitCommit=${GIT_COMMIT}",
	"-X ${PACKAGE}/commands.versionLabel=${VERSION_LABEL}",
}
var ldflagsStr = strings.Join(ldflags, " ")

func init() {
	if specifiedGoExe := os.Getenv("GOEXE"); specifiedGoExe != "" {
		goExe = specifiedGoExe
	}

	if specifiedCliExe := os.Getenv("CLIEXE"); specifiedCliExe != "" {
		cliExe = specifiedCliExe
	}

	// We want to use Go 1.11 modules even if the source lives inside GOPATH.
	// The default is "auto".
	os.Setenv("GO111MODULE", "on")
}

// Run go vet and flake8
func Lint() error {
	fmt.Println("Running go vet...")
	if err := sh.Run(goExe, "vet", packagePath); err != nil {
		return err
	}

	fmt.Println("Running flake8...")
	if err := sh.Run("python3", "-m", "flake8"); err != nil {
		return err
	}

	return nil
}

// Run unit tests
func Unit() error {
	fmt.Println("Running unit tests...")
	return sh.RunV(goExe, "test", "./cli/...")
}

// Run integration tests
func Integration() error {
	fmt.Println("Running integration tests...")
	return sh.RunV("python3", "-m", "pytest")
}

// Run all tests
func Test() {
	mg.SerialDeps(Lint, Unit, Integration)
}

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	fmt.Println("Building...")
	return sh.RunWith(getBuildEnv(), goExe, "build", "-o", cliExe, "-ldflags", ldflagsStr, packagePath)
}

// Download Tarantool Enterprise to tmp dir
func Sdk() error {
	return nil
}

// Activate Tarantool Enterprise to tmp dir
func ActivateSdk() error {
	return nil
}

// Clean up after yourself
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll(cliExe)
}
