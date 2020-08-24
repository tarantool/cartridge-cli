package commands

import "fmt"

// CREATE
const (
	createNameUsage = `Application name`
	templateUsage   = `Application template name
defaults to cartridge`
	createFromUsage = `Path to the application template`
)

// COMMON
const (
	nameUsage = `Application name
defaults to "package" in the rockspec`
)

// PACK
const (
	versionUsage = `Application version
The default version is determined by
"git describe --tags --long"`

	suffixUsage = `Result file (or image) name suffix`

	unitTemplateUsage = `systemd unit template`

	instUnitTemplateUsage = `Instantiated systemd unit template`

	stateboardUnitTemplateUsage = `Stateboard systemd unit template`

	useDockerUsage = `Forces to build the application in Docker`

	tagUsage = `Tag(s) of the result Docker image`

	fromUsage = `Base runtime image Dockerfile
defaults to Dockerfile.cartridge`

	buildFromUsage = `Base build image Dockerfile
defaults to Dockerfile.build.cartridge`

	noCacheUsage = `Use "--no-cache" docker flag
on creation build and runtime images`

	cacheFromUsage = `Use "--cache-from" docker flag
on creation build and runtime images`

	sdkPathUsage = `Path to the SDK to be delivered
defaults to "TARANTOOL_SDK_PATH" env`

	sdkLocalUsage = `Deliver the SDK from the local machine`
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
defaults to "init.lua" ("script" in .cartridge.yml)`

	runDirUsage = `Directory where PID and socket files are stored
defaults to ./tmp/run ("run-dir" in .cartridge.yml)`

	dataDirUsage = `Directory where instances data is stored
defaults to ./tmp/data ("data-dir" in .cartridge.yml)`

	logDirUsage = `Directory where instances logs are stored
defaults to ./tmp/log ("log-dir" in .cartridge.yml)`

	cfgUsage = `Configuration file for instances
defaults to ./instances.yml ("cfg" in .cartridge.yml)`

	daemonizeUsage = `Start instance(s) in background`

	stateboardUsage = `Manage application stateboard as well as instances`

	stateboardOnlyUsage = `Manage only application stateboard`

	logFollowUsage = `Output appended data as the log grows`

	stopForceUsage = `Force instance(s) stop (sends SIGKILL)`
)

// REPAIR
const (
	repairDataDirUsage = `Directory where instances data is stored
Defaults to /var/lib/tarantool
`

	dryRunUsage = `Run command in dry-run mode
Show changes but don't apply them
`

	repairForceUsage = `Repair different configs separately`
)

var (
	timeoutUsage = fmt.Sprintf(`Time to wait for instance(s) start
defaults to %s`, defaultStartTimeout.String())

	logLinesUsage = fmt.Sprintf(`Count of last lines to output
defaults to %d`, defaultLogLines)
)
