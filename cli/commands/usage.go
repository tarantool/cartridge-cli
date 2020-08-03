package commands

import "fmt"

// CREATE
const (
	createNameUsage = `Application name`
	templateUsage   = `Application template`
)

// COMMON
const (
	nameUsage = `Application name
The default name comes from the "package"
field in the rockspec file
`
)

// PACK
const (
	versionUsage = `Application version
The default version is determined by
"git describe --tags --long"
`

	suffixUsage = `Result file (or image) name suffix
`

	unitTemplateUsage = `Path to the template for systemd
unit file
Used for rpm and deb types
`

	instUnitTemplateUsage = `Path to the template for systemd
instantiated unit file
Used for rpm and deb types
`

	stateboardUnitTemplateUsage = `Path to the template for
stateboard systemd unit file
Used for rpm and deb types
`

	useDockerUsage = `Forces to build the application in Docker
`

	tagUsage = `Tag(s) of the Docker image that results
from "pack docker"
Used for docker type
`

	fromUsage = `Path to the base Dockerfile of the runtime
image
Defaults to Dockerfile.cartridge
Used for docker type
`

	buildFromUsage = `Path to the base dockerfile fof the build
image
Used on build in docker
Defaults to Dockerfile.build.cartridge
`

	noCacheUsage = `Creates build and runtime images with
"--no-cache" docker flag
`

	cacheFromUsage = `Images to consider as cache sources
for both build and runtime images
See "--cache-from" docker flag
`

	sdkPathUsage = `Path to the SDK to be delivered
in the result artifact
Alternatively, you can pass the path via the
"TARANTOOL_SDK_PATH" environment variable
`

	sdkLocalUsage = `Flag that indicates if SDK from the local
machine should be delivered in the
result artifact
`
)

// RUNNING
const (
	runningCommonUsage = `Application from current directory is used.
Application name is taken from rockspec in the current directory.

If INSTANCE_NAMEs aren't specified, then all instances described in
config file (see --cfg) are used.

Some flags default options can be override in ./.cartridge.yml config file.
`

	scriptUsage = `Application's entry point
It should be a relative path to the entry point
in the project directory or an absolute path.
Defaults to "init.lua" (or "script" in .cartridge.yml)
`

	runDirUsage = `Directory where PID and socket files are stored
Defaults to ./tmp/run (or "run-dir" in .cartridge.yml)
`

	dataDirUsage = `Directory where instances' data is stored
Each instance's working directory is
"<data-dir>/<app-name>.<instance-name>".
Defaults to ./tmp/data (or "data-dir" in .cartridge.yml)
`

	logDirUsage = `Directory to store instances logs
when running in background
Defaults to ./tmp/log (or "log-dir" in .cartridge.yml)
`

	cfgUsage = `Configuration file for Cartridge instances
Defaults to ./instances.yml (or "cfg" in .cartridge.yml)
`

	daemonizeUsage = `Start instance(s) in background
`

	stateboardUsage = `Manage application stateboard as well as instances
Ignored if "--stateboard-only" is specified
`

	stateboardOnlyUsage = `Manage only application stateboard
If specified, "INSTANCE_NAME..." are ignored.
`

	logFollowUsage = `Output appended data as the log grows
`

	stopForceUsage = `Force instance(s) stop (sends SIGKILL)
`

	createFromUsage = `Path to the application template
`
)

var (
	timeoutUsage = fmt.Sprintf(`Time to wait for instance(s) start
in background
Can be specified in seconds or in duration format
Timeout can't be negative
Timeout 0s means no timeout
Defaults to %s
`, defaultStartTimeout.String())

	logLinesUsage = fmt.Sprintf(`Count of last lines to output
Defaults to %d
`, defaultLogLines)
)
