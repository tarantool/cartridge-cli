// +build mage

package main

import (
	"fmt"
	"io"
	"net/http"
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

var packageName = "github.com/tarantool/cartridge-cli/cli"
var packagePath = "./cli"

var tmpPath = "./tmp"
var sdkDirName = "tarantool-enterprise"
var sdkDirPath = filepath.Join(tmpPath, sdkDirName)

func getBuildEnv() map[string]string {
	var gitTag string
	var gitCommit string

	if _, err := exec.LookPath("git"); err == nil {
		gitTag, _ = sh.Output("git", "describe", "--tags")
		gitCommit, _ = sh.Output("git", "rev-parse", "--short", "HEAD")

	}

	versionLabel := os.Getenv("VERSION_LABEL")

	return map[string]string{
		"PACKAGE":       packageName,
		"GIT_TAG":       gitTag,
		"GIT_COMMIT":    gitCommit,
		"VERSION_LABEL": versionLabel,
	}
}

var ldflags = []string{
	"-s", "-w",
	"-X ${PACKAGE}/version.gitTag=${GIT_TAG}",
	"-X ${PACKAGE}/version.gitCommit=${GIT_COMMIT}",
	"-X ${PACKAGE}/version.versionLabel=${VERSION_LABEL}",
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
	if err := sh.RunV(goExe, "vet", packagePath); err != nil {
		return err
	}

	fmt.Println("Running flake8...")
	if err := sh.RunV(py3Exe, "-m", "flake8"); err != nil {
		return err
	}

	return nil
}

// Run unit tests
func Unit() error {
	fmt.Println("Running unit tests...")
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

// A build step that requires additional params, or platform specific steps for example
func Build() error {
	fmt.Println("Building...")
	return sh.RunWith(
		getBuildEnv(), goExe, "build",
		"-o", cliExe,
		"-ldflags", ldflagsStr,
		packagePath,
	)
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
	os.RemoveAll(cliExe)
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

	archivedSDKPath := filepath.Join(tmpPath, archivedSDKName)
	archivedSDKFile, err := os.Create(archivedSDKPath)
	if err != nil {
		return fmt.Errorf("Failed to create archived SDK file: %s", err)
	}
	defer archivedSDKFile.Close()
	defer os.RemoveAll(archivedSDKFile.Name())

	fmt.Printf("Download Tarantool Enterprise SDK %s...\n", bundleVersion)

	resp, err := http.Get(sdkDownloadUrl)
	if err != nil {
		return fmt.Errorf("Failed to download archived SDK: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to download archived SDK: %s", resp.Status)
	}

	if _, err := io.Copy(archivedSDKFile, resp.Body); err != nil {
		return fmt.Errorf("Failed to download archived SDK: %s", err)
	}

	fmt.Println("Unarchive Tarantool Enterprise SDK...")

	if err := sh.RunV("tar", "-xzf", archivedSDKPath, "-C", tmpPath); err != nil {
		return fmt.Errorf("Failed to unarchive SDK: %s")
	}

	return nil
}
