package pack

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/tarantool/cartridge-cli/src/common"
	"github.com/tarantool/cartridge-cli/src/project"
)

var (
	extByType = map[string]string{
		tgzType: "tar.gz",
		rpmType: "rpm",
		debType: "deb",
	}

	versionRgxps = []*regexp.Regexp{
		regexp.MustCompile(`^(?P<Major>\d+)$`),
		regexp.MustCompile(`^(?P<Major>\d+)\.(?P<Minor>\d+)$`),
		regexp.MustCompile(`^(?P<Major>\d+)\.(?P<Minor>\d+)\.(?P<Patch>\d+)$`),
		regexp.MustCompile(`^(?P<Major>\d+)\.(?P<Minor>\d+)\.(?P<Patch>\d+)-(?P<Count>\d+)$`),
		regexp.MustCompile(`^(?P<Major>\d+)\.(?P<Minor>\d+)\.(?P<Patch>\d+)-(?P<Hash>\w+)$`),
		regexp.MustCompile(
			`^(?P<Major>\d+)\.(?P<Minor>\d+)\.(?P<Patch>\d+)-(?P<Count>\d+)-(?P<Hash>\w+)$`,
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
		// TODO: handle internal error
		panic(fmt.Errorf("Unknown type: %s", projectCtx.PackType))
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

func getImageFullname(projectCtx *project.ProjectCtx) string {
	var imageFullname string

	if projectCtx.ImageTag != "" {
		imageFullname = projectCtx.ImageTag
	} else {
		imageFullname = fmt.Sprintf(
			"%s:%s",
			projectCtx.Name,
			projectCtx.VersionRelease,
		)

		if projectCtx.Suffix != "" {
			imageFullname = fmt.Sprintf(
				"%s-%s",
				imageFullname,
				projectCtx.Suffix,
			)
		}
	}

	return imageFullname
}

func checkTagVersionSuffix(projectCtx *project.ProjectCtx) error {
	if projectCtx.PackType != dockerType {
		if projectCtx.ImageTag != "" {
			log.Warnf(tagForNonDockerWarn)
		}
		return nil
	}

	if projectCtx.ImageTag != "" && (projectCtx.Version != "" || projectCtx.Suffix != "") {
		return fmt.Errorf(tagVersionSuffixErr)
	}

	return nil
}

const (
	tagForNonDockerWarn = `Ignored --tag option specific for docker type`
	tagVersionSuffixErr = `You can specify only --version (and --suffix) or --tag options`
)
