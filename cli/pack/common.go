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

	// Check if --version or --suffix is allowed to be added in package version.
	packageVerAllowed = map[string]*regexp.Regexp{
		// The version number is normally taken verbatim from the package's version.
		// The only restriction placed on the version is that it cannot contain a dash "-".
		// http://ftp.rpm.org/max-rpm/ch-rpm-file-format.html
		RpmType: regexp.MustCompile(`^[^-]+$`),
		// The format is: [epoch:]upstream_version[-debian_revision].
		// The upstream_version must contain only alphanumerics 6
		// and the characters . + - ~ (full stop, plus, hyphen, tilde)
		// and should start with a digit. If there is no debian_revision
		// then hyphens are not allowed (we have one).
		// https://www.debian.org/doc/debian-policy/ch-controlfields.html#version
		// https://www.debian.org/doc/manuals/debian-reference/ch02.en.html#_debian_package_file_names
		DebType: regexp.MustCompile(`^[-a-zA-Z0-9.+:~]+$`),
	}
)

func normalizeGitVersion(ctx *context.Ctx) error {
	var major = "0"
	var minor = "0"
	var patch = "0"
	var count = ""

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
				}
			}
			break
		}
	}

	if !matched {
		return fmt.Errorf("Git tag should be semantic (major.minor.patch)")
	}

	if count == "" {
		count = "0"
	}

	ctx.Pack.Version = fmt.Sprintf("%s.%s.%s.%s", major, minor, patch, count)

	return nil
}

func buildVersionWithSuffix(ctx *context.Ctx) error {
	suffix := strings.TrimSpace(ctx.Pack.Suffix)

	if suffix == "" {
		ctx.Pack.VersionWithSuffix = ctx.Pack.Version
		return nil
	}

	switch ctx.Pack.Type {
	case RpmType:
		if !packageVerAllowed[RpmType].MatchString(suffix) {
			return fmt.Errorf("Dashes are not allowed in RPM --suffix")
		}
	case DebType:
		if !packageVerAllowed[DebType].MatchString(suffix) {
			return fmt.Errorf("DEB --suffix must satisfy [-a-zA-Z0-9.+:~]+")
		}
	}

	ctx.Pack.VersionWithSuffix = fmt.Sprintf("%s.%s", ctx.Pack.Version, suffix)

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

		if err := normalizeGitVersion(ctx); err != nil {
			return err
		}
	} else {
		switch ctx.Pack.Type {
		case RpmType:
			if !packageVerAllowed[RpmType].MatchString(ctx.Pack.Version) {
				return fmt.Errorf("Dashes are not allowed in RPM --version")
			}
		case DebType:
			if !packageVerAllowed[DebType].MatchString(ctx.Pack.Version) {
				return fmt.Errorf("DEB --version must satisfy [-a-zA-Z0-9.+:]+")
			}
		}
	}

	if err := buildVersionWithSuffix(ctx); err != nil {
		return err
	}

	return nil
}

func detectRelease(ctx *context.Ctx) {
	// For DEB package, this part of the version number specifies the version
	// of the Debian package based on the upstream version.
	// It is conventional to restart the debian_revision at 1
	// each time the upstream_version is increased.

	// For RPM package, `release` is the number of times this version
	// of the software has been packaged.

	ctx.Pack.Release = "1"
}

func detectArch(ctx *context.Ctx) {
	switch ctx.Pack.Type {
	case RpmType:
		ctx.Pack.Arch = "x86_64"
	case DebType:
		ctx.Pack.Arch = "all"
	case TgzType:
		ctx.Pack.Arch = "x86_64"
	}
}

func getRpmPackageFullname(ctx *context.Ctx) string {
	// RPM name convention is "name-version-release.architecture.rpm"
	// http://ftp.rpm.org/max-rpm/ch-rpm-file-format.html
	return fmt.Sprintf(
		"%s-%s-%s.%s.%s",
		ctx.Project.Name,
		ctx.Pack.VersionWithSuffix,
		ctx.Pack.Release,
		ctx.Pack.Arch,
		extByType[ctx.Pack.Type],
	)
}

func getDebPackageFullname(ctx *context.Ctx) string {
	// DEB name convention is "package-name_upstream-version-debian.revision_architecture.deb".
	// https://www.debian.org/doc/manuals/debian-reference/ch02.en.html#_debian_package_file_names
	// https://www.debian.org/doc/debian-policy/ch-controlfields.html#version
	return fmt.Sprintf(
		"%s_%s-%s_%s.%s",
		ctx.Project.Name,
		ctx.Pack.VersionWithSuffix,
		ctx.Pack.Release,
		ctx.Pack.Arch,
		extByType[ctx.Pack.Type],
	)
}

func getTgzPackageFullname(ctx *context.Ctx) string {
	return fmt.Sprintf(
		"%s-%s.%s.%s",
		ctx.Project.Name,
		ctx.Pack.VersionWithSuffix,
		ctx.Pack.Arch,
		extByType[ctx.Pack.Type],
	)
}

func getPackageFullname(ctx *context.Ctx) string {
	if _, found := extByType[ctx.Pack.Type]; !found {
		panic(project.InternalError("Unknown type: %s", ctx.Pack.Type))
	}

	switch ctx.Pack.Type {
	case RpmType:
		return getRpmPackageFullname(ctx)
	case DebType:
		return getDebPackageFullname(ctx)
	default:
		return getTgzPackageFullname(ctx)
	}
}

func getImageTags(ctx *context.Ctx) []string {
	var imageTags []string

	if len(ctx.Pack.ImageTags) > 0 {
		imageTags = ctx.Pack.ImageTags
	} else {
		ImageTags := fmt.Sprintf(
			"%s:%s",
			ctx.Project.Name,
			ctx.Pack.Version,
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
