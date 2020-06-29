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

type RepoInfo struct {
	OS   string
	Dist string
}

var targetRepos = []RepoInfo{
	{OS: "el", Dist: "6"},
	{OS: "el", Dist: "7"},
	{OS: "el", Dist: "8"},
	{OS: "fedora", Dist: "29"},
	{OS: "fedora", Dist: "30"},

	{OS: "ubuntu", Dist: "trusty"},
	{OS: "ubuntu", Dist: "xenial"},
	{OS: "ubuntu", Dist: "bionic"},
	{OS: "ubuntu", Dist: "eoan"},
	{OS: "debian", Dist: "jessie"},
	{OS: "debian", Dist: "stretch"},
	{OS: "debian", Dist: "buster"},
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
	defer os.RemoveAll(s3UpdateRepoScriptName)

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

	for _, targetRepo := range targetRepos {
		fmt.Printf("Publish package for %s/%s...\n", targetRepo.OS, targetRepo.Dist)

		err := sh.RunV(
			"bash", publishCtx["s3UpdateRepoScriptPath"],
			fmt.Sprintf("-o=%s", targetRepo.OS),
			fmt.Sprintf("-d=%s", targetRepo.Dist),
			fmt.Sprintf("-p=%s", packageName),
			fmt.Sprintf("-b=%s", publishCtx["s3RepoPath"]),
			fmt.Sprintf("-f"),
			publishCtx["distPath"],
		)

		if err != nil {
			return fmt.Errorf("Failed to publish package for %s/%s: %s", targetRepo.OS, targetRepo.Dist, err)
		}
	}

	return nil
}
