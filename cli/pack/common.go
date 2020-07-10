package pack

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
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

func normalizeVersion(ctx *context.Ctx) error {
	var major = "0"
	var minor = "0"
	var patch = "0"
	var count = ""
	var hash = ""

	matched := false
	for _, r := range versionRgxps {
		matches := r.FindStringSubmatch(ctx.Pack.Version)
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

	ctx.Pack.Version = fmt.Sprintf("%s.%s.%s", major, minor, patch)

	if count != "" && hash != "" {
		ctx.Pack.Release = fmt.Sprintf("%s-%s", count, hash)
	} else if count != "" {
		ctx.Pack.Release = count
	} else if hash != "" {
		ctx.Pack.Release = hash
	} else {
		ctx.Pack.Release = "0"
	}

	ctx.Pack.VersionRelease = fmt.Sprintf("%s-%s", ctx.Pack.Version, ctx.Pack.Release)

	return nil
}

func detectVersion(ctx *context.Ctx) error {
	if ctx.Pack.Version == "" {
		if !common.GitIsInstalled() {
			return fmt.Errorf("git not found. " +
				"Please pass version explicitly via --version")
		} else if !common.IsGitProject(ctx.Project.Path) {
			return fmt.Errorf("Project is not a git project. " +
				"Please pass version explicitly via --version")
		}

		gitDescribeCmd := exec.Command("git", "describe", "--tags", "--long")
		gitVersion, err := common.GetOutput(gitDescribeCmd, &ctx.Project.Path)

		if err != nil {
			return fmt.Errorf("Failed to get version using git: %s", err)
		}

		ctx.Pack.Version = strings.Trim(gitVersion, "\n")
	}

	if err := normalizeVersion(ctx); err != nil {
		return err
	}

	return nil
}

func getPackageFullname(ctx *context.Ctx) string {
	ext, found := extByType[ctx.Pack.Type]
	if !found {
		panic(project.InternalError("Unknown type: %s", ctx.Pack.Type))
	}

	packageFullname := fmt.Sprintf(
		"%s-%s",
		ctx.Project.Name,
		ctx.Pack.VersionRelease,
	)

	if ctx.Pack.Suffix != "" {
		packageFullname = fmt.Sprintf(
			"%s-%s",
			packageFullname,
			ctx.Pack.Suffix,
		)
	}

	packageFullname = fmt.Sprintf(
		"%s.%s",
		packageFullname,
		ext,
	)

	return packageFullname
}

func getImageTags(ctx *context.Ctx) []string {
	var imageTags []string

	if len(ctx.Pack.ImageTags) > 0 {
		imageTags = ctx.Pack.ImageTags
	} else {
		ImageTags := fmt.Sprintf(
			"%s:%s",
			ctx.Project.Name,
			ctx.Pack.VersionRelease,
		)

		if ctx.Pack.Suffix != "" {
			ImageTags = fmt.Sprintf(
				"%s-%s",
				ImageTags,
				ctx.Pack.Suffix,
			)
		}

		imageTags = []string{ImageTags}
	}

	return imageTags
}

func checkTagVersionSuffix(ctx *context.Ctx) error {
	if ctx.Pack.Type != DockerType {
		return nil
	}

	if len(ctx.Pack.ImageTags) > 0 && (ctx.Pack.Version != "" || ctx.Pack.Suffix != "") {
		return fmt.Errorf(tagVersionSuffixErr)
	}

	return nil
}

const (
	tagVersionSuffixErr = `You can specify only --version (and --suffix) or --tag options`
)
