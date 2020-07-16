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
	runningCommonUsage = `Manage instance(s) of current application

There are two modes of running instances: local and global.

[local]: Instances of application in the current dir are managed .
Application name is taken from rockspec in the current directory.
./.cartridge.yml is used to read default options values.

[global]: Instances of application from <apps-dir>/<name> are managed.
Application name should be specified via --name or as a first argument (APP_NAME).
~/.cartridge.yml is used to read default options values.

If INSTANCE_NAMEs aren't specified, then all instances described in
config file (see --cfg) are used.
`

	globalFlagDoc = `Manage instance(s) globally
Name of application to manage should be specified
as a first argument APP_NAME or via --name option
`

	appsDirUsage = `Directory where applications are stored
Is used only for global running
Defaults to [global] /usr/share/tarantool
            [.cartridge.yml] "apps-dir"
`

	scriptUsage = `Application's entry point
Relative to the application directory or absolute
Defaults to "init.lua"
            [.cartridge.yml] "script"
`

	runDirUsage = `Directory where PID and socket files are stored
Defaults to [local] ./tmp/run
            [global] /var/run/tarantool
            [.cartridge.yml] "run-dir"
`

	dataDirUsage = `Directory where instances' data is stored
Defaults to [local] ./tmp/data
            [global] /var/lib/tarantool
            [.cartridge.yml] "data-dir"
`

	logDirUsage = `Directory to store instances' logs
when running in background
Defaults to [local] ./tmp/log
            [global] /var/log/tarantool
            [.cartridge.yml] "log-dir"
`

	appConfUsage = `Configuration for Cartridge instances
Defaults to [local] ./instances.yml
            [global] /etc/tarantool/conf.d
            [.cartridge.yml] "cfg"
`

	daemonizeUsage = `Start instance(s) in background
`

	stateboardUsage = `Manage application stateboard as well as instances
Ignored if "--stateboard-only" is specified
`

	stateboardOnlyUsage = `Manage only application stateboard
If specified, "INSTANCE_NAME..." are ignored
`

	logFollowUsage = `Output appended data as the log grows
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
