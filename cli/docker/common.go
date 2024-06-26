package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	goVersion "github.com/hashicorp/go-version"
)

var (
	dockerServerMinVersion *goVersion.Version
)

func init() {
	dockerServerMinVersion = goVersion.Must(goVersion.NewSemver("17.03.2"))
}

func getServerVersion() (string, error) {
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}

	ctx := context.Background()
	version, err := cli.ServerVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("Failed to get docker server version: %s", err)
	}

	return version.Version, nil
}

func CheckMinServerVersion() error {
	serverVersionStr, err := getServerVersion()
	if err != nil {
		return fmt.Errorf("Failed to check docker server version: %s", err)
	}

	serverVersion, err := goVersion.NewSemver(serverVersionStr)
	if err != nil {
		return fmt.Errorf("Failed to parse docker server version: %s", err)
	}

	if serverVersion.LessThan(dockerServerMinVersion) {
		return fmt.Errorf(
			"Docker version %s is not supported. Minimal required docker version is %s",
			serverVersion, dockerServerMinVersion,
		)

	}

	return nil
}
