package docker

import (
	"context"
	"fmt"

	client "docker.io/go-docker"
)

const (
	dockerServerMinVersion = "19.03.8" // TODO: test on v18
)

func getServerVersion() (string, error) {
	cli, err := client.NewEnvClient()
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
	serverVersion, err := getServerVersion()
	if err != nil {
		return fmt.Errorf("Failed to check docker server version: %s", err)
	}

	if serverVersion < dockerServerMinVersion {
		return fmt.Errorf(
			"Docker version %s is not supported. Minimal required docker version is %s",
			serverVersion, dockerServerMinVersion,
		)

	}

	return nil
}
