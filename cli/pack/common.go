package pack

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/project"
)

var (
	extByType = map[string]string{
		TgzType: "tar.gz",
		RpmType: "rpm",
		DebType: "deb",
	}

	versionRgxps = []*regexp.Regexp{
		regexp.MustCompile(`^(?P<Major>\d+)$`),
		regexp.MustCompile(`^(?P<Major>\d+)\.(?P<Minor>\d+)$`),
		regexp.MustCompile(`^(?P<Major>\d+)\.(?P<Minor>\d+)\.(?P<Patch>\d+)$`),
		regexp.MustCompile(`^(?P<Major>\d+)\.(?P<Minor>\d+)\.(?P<Patch>\d+)-(?P<Count>\d+)$`),
		regexp.MustCompile(`^(?P<Major>\d+)\.(?P<Minor>\d+)\.(?P<Patch>\d+)-(?P<Hash>g\w+)$`),
		regexp.MustCompile(
			`^(?P<Major>\d+)\.(?P<Minor>\d+)\.(?P<Patch>\d+)-(?P<Count>\d+)-(?P<Hash>g\w+)$`,
		),
	}
)

func normalizeVersion(projectCtx *project.ProjectCtx) error {
	var major = "0"
	var minor = "0"
	var patch = "0"
	var count = ""
	var hash = ""

	matched := false
	for _, r := range versionRgxps {
		matches := r.FindStringSubmatch(projectCtx.Version)
		if matches != nil {
			matched = true
			for i, expName := range r.SubexpNames() {
				switch expName {
				case "Major":
					major = matches[i]
				case "Minor":
					minor = matches[i]
				case "Patch":
					patch = matches[i]
				case "Count":
					count = matches[i]
				case "Hash":
					hash = matches[i]
				}
			}
			break
		}
	}

	if !matched {
		return fmt.Errorf("Version should be semantic (major.minor.patch[-count][-commit])")
	}

	projectCtx.Version = fmt.Sprintf("%s.%s.%s", major, minor, patch)

	if count != "" && hash != "" {
		projectCtx.Release = fmt.Sprintf("%s-%s", count, hash)
	} else if count != "" {
		projectCtx.Release = count
	} else if hash != "" {
		projectCtx.Release = hash
	} else {
		projectCtx.Release = "0"
	}

	projectCtx.VersionRelease = fmt.Sprintf("%s-%s", projectCtx.Version, projectCtx.Release)

	return nil
}

func detectVersion(projectCtx *project.ProjectCtx) error {
	if projectCtx.Version == "" {
		if !common.GitIsInstalled() {
			return fmt.Errorf("git not found. " +
				"Please pass version explicitly via --version")
		} else if !common.IsGitProject(projectCtx.Path) {
			return fmt.Errorf("Project is not a git project. " +
				"Please pass version explicitly via --version")
		}

		gitDescribeCmd := exec.Command("git", "describe", "--tags", "--long")
		gitVersion, err := common.GetOutput(gitDescribeCmd, &projectCtx.Path)

		if err != nil {
			return fmt.Errorf("Failed to get version using git: %s", err)
		}

		projectCtx.Version = strings.Trim(gitVersion, "\n")
	}

	if err := normalizeVersion(projectCtx); err != nil {
		return err
	}

	return nil
}

func getPackageFullname(projectCtx *project.ProjectCtx) string {
	ext, found := extByType[projectCtx.PackType]
	if !found {
		panic(project.InternalError("Unknown type: %s", projectCtx.PackType))
	}

	packageFullname := fmt.Sprintf(
		"%s-%s",
		projectCtx.Name,
		projectCtx.VersionRelease,
	)

	if projectCtx.Suffix != "" {
		packageFullname = fmt.Sprintf(
			"%s-%s",
			packageFullname,
			projectCtx.Suffix,
		)
	}

	packageFullname = fmt.Sprintf(
		"%s.%s",
		packageFullname,
		ext,
	)

	return packageFullname
}

func getImageTags(projectCtx *project.ProjectCtx) []string {
	var imageTags []string

	if len(projectCtx.ImageTags) > 0 {
		imageTags = projectCtx.ImageTags
	} else {
		ImageTags := fmt.Sprintf(
			"%s:%s",
			projectCtx.Name,
			projectCtx.VersionRelease,
		)

		if projectCtx.Suffix != "" {
			ImageTags = fmt.Sprintf(
				"%s-%s",
				ImageTags,
				projectCtx.Suffix,
			)
		}

		imageTags = []string{ImageTags}
	}

	return imageTags
}

func checkTagVersionSuffix(projectCtx *project.ProjectCtx) error {
	if projectCtx.PackType != DockerType {
		return nil
	}

	if len(projectCtx.ImageTags) > 0 && (projectCtx.Version != "" || projectCtx.Suffix != "") {
		return fmt.Errorf(tagVersionSuffixErr)
	}

	return nil
}

const (
	tagVersionSuffixErr = `You can specify only --version (and --suffix) or --tag options`
)
