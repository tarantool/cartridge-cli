// +build mage

package main

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/magefile/mage/sh"
)

const s3UpdateRepoScriptEnv = "S3_UPDATE_REPO_SCRIPT_URL"
const s3UpdateRepoScriptName = "update_repo.sh"
const s3BucketEnv = "S3_BUCKET"
const s3FolderEnv = "S3_FOLDER"

const distPath = "dist"

const packageName = "cartridge-cli"
const rpmExt = "rpm"
const debExt = "deb"

type Distro struct {
	OS   string
	Dist string
}

var targetDistros = []Distro{
	{OS: "el", Dist: "6"},
	{OS: "el", Dist: "7"},
	{OS: "el", Dist: "8"},
	{OS: "fedora", Dist: "29"},
	{OS: "fedora", Dist: "30"},

	{OS: "ubuntu", Dist: "trusty"},
	{OS: "ubuntu", Dist: "xenial"},
	{OS: "ubuntu", Dist: "bionic"},
	{OS: "ubuntu", Dist: "eoan"},
	{OS: "ubuntu", Dist: "focal"},

	{OS: "debian", Dist: "jessie"},
	{OS: "debian", Dist: "stretch"},
	{OS: "debian", Dist: "buster"},
	{OS: "debian", Dist: "bullseye"},
}

func getArch(distro Distro) (string, error) {
	if distro.OS == "el" || distro.OS == "fedora" {
		return "x86_64", nil
	}

	if distro.OS == "ubuntu" || distro.OS == "debian" {
		return "amd64", nil
	}

	return "", fmt.Errorf("Unknown OS: %s", distro.OS)
}

func getExt(distro Distro) (string, error) {
	if distro.OS == "el" || distro.OS == "fedora" {
		return rpmExt, nil
	}

	if distro.OS == "ubuntu" || distro.OS == "debian" {
		return debExt, nil
	}

	return "", fmt.Errorf("Unknown OS: %s", distro.OS)
}

func getPackagePath(distro Distro) (string, error) {
	ext, err := getExt(distro)
	if err != nil {
		return "", fmt.Errorf("Failed to get ext: %s", err)
	}

	arch, err := getArch(distro)
	if err != nil {
		return "", fmt.Errorf("Failed to get arch: %s", err)
	}

	var packageNamePattern string
	if ext == "deb" {
		packageNamePattern = fmt.Sprintf("%s_*_%s.deb", packageName, arch)
	} else if ext == "rpm" {
		packageNamePattern = fmt.Sprintf("%s-*.%s.rpm", packageName, arch)
	} else {
		return "", fmt.Errorf("Unknown extension: %s", ext)
	}

	matches, err := filepath.Glob(filepath.Join(distPath, packageNamePattern))
	if err != nil {
		return "", fmt.Errorf("Failed to find matched files: %s", err)
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("No matched packages found for %s", packageNamePattern)
	} else if len(matches) > 1 {
		return "", fmt.Errorf("Found multiple matched packages for %s: %v", packageNamePattern, matches)
	}

	return matches[0], nil
}

func getS3Ctx() (map[string]string, error) {
	s3Bucket := os.Getenv(s3BucketEnv)
	if s3Bucket == "" {
		return nil, fmt.Errorf("Please, specify %s", s3BucketEnv)
	}

	s3Folder := os.Getenv(s3FolderEnv)
	if s3Folder == "" {
		return nil, fmt.Errorf("Please, specify %s", s3FolderEnv)
	}

	s3UpdateRepoScriptUrl := os.Getenv(s3UpdateRepoScriptEnv)
	if s3UpdateRepoScriptUrl == "" {
		return nil, fmt.Errorf("Please, specify %s", s3UpdateRepoScriptEnv)
	}

	s3UpdateRepoScriptPath := filepath.Join(tmpPath, s3UpdateRepoScriptName)
	if err := downloadFile(s3UpdateRepoScriptUrl, s3UpdateRepoScriptPath); err != nil {
		return nil, fmt.Errorf("Failed to download update S3 repo script: %s", err)
	}

	s3BucketURL, err := url.Parse(s3Bucket)
	if err != nil {
		return nil, fmt.Errorf("Invalid S3 bucket URL passed: %s", err)
	}
	s3BucketURL.Path = path.Join(s3BucketURL.Path, s3Folder)
	s3RepoPath := s3BucketURL.String()

	s3Ctx := map[string]string{
		"distPath":               distPath,
		"s3RepoPath":             s3RepoPath,
		"s3UpdateRepoScriptPath": s3UpdateRepoScriptPath,
	}

	return s3Ctx, nil
}

// publish RPM and DEB packages to S3
func PublishS3() error {
	fmt.Printf("Publish packages to S3...\n")

	publishCtx, err := getS3Ctx()
	defer os.RemoveAll(publishCtx["s3UpdateRepoScriptPath"])
	if err != nil {
		return err
	}

	for _, targetDistro := range targetDistros {
		fmt.Printf("Publish package for %s/%s...\n", targetDistro.OS, targetDistro.Dist)

		err := sh.RunV(
			"bash", publishCtx["s3UpdateRepoScriptPath"],
			fmt.Sprintf("-o=%s", targetDistro.OS),
			fmt.Sprintf("-d=%s", targetDistro.Dist),
			fmt.Sprintf("-p=%s", packageName),
			fmt.Sprintf("-b=%s", publishCtx["s3RepoPath"]),
			fmt.Sprintf("-f"),
			publishCtx["distPath"],
		)

		if err != nil {
			return fmt.Errorf("Failed to publish package for %s/%s: %s", targetDistro.OS, targetDistro.Dist, err)
		}
	}

	return nil
}
