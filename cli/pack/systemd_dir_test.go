package pack

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarantool/cartridge-cli/cli/context"
)

func writeSystemdUnitParams(file *os.File, content string) {
	if err := ioutil.WriteFile(file.Name(), []byte(content), 0644); err != nil {
		panic(fmt.Errorf("Failed to write systemd unit params: %s", err))
	}
}

func TestCheckBaseUnitFiles(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	var err error

	// fill Ctx
	var ctx context.Ctx

	// unit file args
	ctx.Project.Name = "test-app"
	ctx.Running.DataDir = "/var/lib/tarantool/"
	ctx.Running.ConfPath = "/etc/tarantool/conf.d"
	ctx.Running.RunDir = "/var/run/tarantool/"

	// stateboard unit file args
	ctx.Project.StateboardName = "test-app-stateboard"
	ctx.Running.WithStateboard = true

	expUnitContent := `[Unit]
Description=Tarantool Cartridge app test-app.default
After=network.target

[Service]
Type=simple
ExecStartPre=/bin/sh -c 'mkdir -p /var/lib/tarantool/test-app.default'
ExecStart=/usr/bin/tarantool 
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool

Environment=TARANTOOL_APP_NAME=test-app
Environment=TARANTOOL_CFG=/etc/tarantool/conf.d
Environment=TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/test-app.default.control
Environment=TARANTOOL_NET_MSG_MAX=768
Environment=TARANTOOL_PID_FILE=/var/run/tarantool/test-app.default.pid
Environment=TARANTOOL_WORKDIR=/var/lib/tarantool/test-app.default


LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Increase fd limit for Vinyl
LimitNOFILE=65535

# Systemd waits until all xlogs are recovered
TimeoutStartSec=86400s
# Give a reasonable amount of time to close xlogs
TimeoutStopSec=10s

[Install]
WantedBy=multi-user.target
Alias=test-app
`

    expInstUnitContent := `[Unit]
Description=Tarantool Cartridge app test-app@%i
After=network.target

[Service]
Type=simple
ExecStartPre=/bin/sh -c 'mkdir -p /var/lib/tarantool/test-app.%i'
ExecStart=/usr/bin/tarantool 
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool

Environment=TARANTOOL_APP_NAME=test-app
Environment=TARANTOOL_CFG=/etc/tarantool/conf.d
Environment=TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/test-app.%i.control
Environment=TARANTOOL_INSTANCE_NAME=%i
Environment=TARANTOOL_NET_MSG_MAX=768
Environment=TARANTOOL_PID_FILE=/var/run/tarantool/test-app.%i.pid
Environment=TARANTOOL_WORKDIR=/var/lib/tarantool/test-app.%i


LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Increase fd limit for Vinyl
LimitNOFILE=65535

# Systemd waits until all xlogs are recovered
TimeoutStartSec=86400s
# Give a reasonable amount of time to close xlogs
TimeoutStopSec=10s

[Install]
WantedBy=multi-user.target
Alias=test-app.%i
`
    expStateboardUnitContent := `[Unit]
Description=Tarantool Cartridge stateboard for test-app
After=network.target

[Service]
Type=simple
ExecStartPre=/bin/sh -c 'mkdir -p /var/lib/tarantool/test-app-stateboard'
ExecStart=/usr/bin/tarantool 
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool

Environment=TARANTOOL_APP_NAME=test-app-stateboard
Environment=TARANTOOL_CFG=/etc/tarantool/conf.d
Environment=TARANTOOL_CONSOLE_SOCK=/var/run/tarantool/test-app-stateboard.control
Environment=TARANTOOL_NET_MSG_MAX=768
Environment=TARANTOOL_PID_FILE=/var/run/tarantool/test-app-stateboard.pid
Environment=TARANTOOL_WORKDIR=/var/lib/tarantool/test-app-stateboard


LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Increase fd limit
LimitNOFILE=65535

# Systemd waits until all xlogs are recovered
TimeoutStartSec=86400s
# Give a reasonable amount of time to close xlogs
TimeoutStopSec=10s

[Install]
WantedBy=multi-user.target
Alias=test-app-stateboard
`
    expContentByFilename := map[string]string{
		fmt.Sprintf("%s.service", ctx.Project.Name): expUnitContent,
		fmt.Sprintf("%s@.service", ctx.Project.Name): expInstUnitContent,
		fmt.Sprintf("%s.service", ctx.Project.StateboardName): expStateboardUnitContent,
	}

	// create tmp directory
	tmpDir, err := ioutil.TempDir("", "tmp")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	err = initSystemdDir(tmpDir, &ctx)
	assert.Nil(err)

	// read systemd directory content
	systemdDirPath := filepath.Join(tmpDir, "/etc/systemd/system/")

	for filename, expContent := range expContentByFilename {
		content, err := ioutil.ReadFile(filepath.Join(systemdDirPath, filename))
		if err != nil {
			log.Fatal(err)
		}
		assert.Equal(string(content), expContent)
	}
}

func TestCheckSpecifiedArgsUnitFiles(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	var err error
	var ctx context.Ctx

	newAppName := "new-name"
	newStateboardAppName := "new-stateboard-name"

	systemdUnitParamsContent := fmt.Sprintf(`fd-limit: 1024
stateboard-fd-limit: 2048
instance-env:
    app-name: %s
    net-msg-max: 2048
    workdir: /new/workdir/
    pid-file: /new/path/to/pidfile/
    console-sock: /new/path/to/console/sock/
    cfg: /new/path/to/cfg/
    user-param: my-param
stateboard-env:
    app-name: %s
    net-msg-max: 1024
    workdir: /new/stateboard/workdir/
    pid-file: /new/stateboard/path/to/pidfile/
    console-sock: /new/stateboard/path/to/console/sock/
    cfg: /new/stateboard/path/to/cfg/
    user-stateboard-param: my-stateboard-param
`, newAppName, newStateboardAppName)

	// create tmp systemd unit params file
	f, err := ioutil.TempFile("", "systemd-unit-params*.yml")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	writeSystemdUnitParams(f, systemdUnitParamsContent)

	// fill ctx
	ctx.Running.WithStateboard = true
	ctx.Pack.SystemdUnitParamsPath = f.Name()

	expUnitContent := `[Unit]
Description=Tarantool Cartridge app new-name.default
After=network.target

[Service]
Type=simple
ExecStartPre=/bin/sh -c 'mkdir -p /new/workdir/'
ExecStart=/usr/bin/tarantool 
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool

Environment=TARANTOOL_APP_NAME=new-name
Environment=TARANTOOL_CFG=/new/path/to/cfg/
Environment=TARANTOOL_CONSOLE_SOCK=/new/path/to/console/sock/
Environment=TARANTOOL_NET_MSG_MAX=2048
Environment=TARANTOOL_PID_FILE=/new/path/to/pidfile/
Environment=TARANTOOL_USER_PARAM=my-param
Environment=TARANTOOL_WORKDIR=/new/workdir/


LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Increase fd limit for Vinyl
LimitNOFILE=1024

# Systemd waits until all xlogs are recovered
TimeoutStartSec=86400s
# Give a reasonable amount of time to close xlogs
TimeoutStopSec=10s

[Install]
WantedBy=multi-user.target
Alias=new-name
`

	expInstUnitContent := `[Unit]
Description=Tarantool Cartridge app new-name@%i
After=network.target

[Service]
Type=simple
ExecStartPre=/bin/sh -c 'mkdir -p /new/workdir/'
ExecStart=/usr/bin/tarantool 
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool

Environment=TARANTOOL_APP_NAME=new-name
Environment=TARANTOOL_CFG=/new/path/to/cfg/
Environment=TARANTOOL_CONSOLE_SOCK=/new/path/to/console/sock/
Environment=TARANTOOL_INSTANCE_NAME=%i
Environment=TARANTOOL_NET_MSG_MAX=2048
Environment=TARANTOOL_PID_FILE=/new/path/to/pidfile/
Environment=TARANTOOL_USER_PARAM=my-param
Environment=TARANTOOL_WORKDIR=/new/workdir/


LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Increase fd limit for Vinyl
LimitNOFILE=1024

# Systemd waits until all xlogs are recovered
TimeoutStartSec=86400s
# Give a reasonable amount of time to close xlogs
TimeoutStopSec=10s

[Install]
WantedBy=multi-user.target
Alias=new-name.%i
`
	expStateboardUnitContent := `[Unit]
Description=Tarantool Cartridge stateboard for new-name
After=network.target

[Service]
Type=simple
ExecStartPre=/bin/sh -c 'mkdir -p /new/stateboard/workdir/'
ExecStart=/usr/bin/tarantool 
Restart=on-failure
RestartSec=2
User=tarantool
Group=tarantool

Environment=TARANTOOL_APP_NAME=new-stateboard-name
Environment=TARANTOOL_CFG=/new/stateboard/path/to/cfg/
Environment=TARANTOOL_CONSOLE_SOCK=/new/stateboard/path/to/console/sock/
Environment=TARANTOOL_NET_MSG_MAX=1024
Environment=TARANTOOL_PID_FILE=/new/stateboard/path/to/pidfile/
Environment=TARANTOOL_USER_STATEBOARD_PARAM=my-stateboard-param
Environment=TARANTOOL_WORKDIR=/new/stateboard/workdir/


LimitCORE=infinity
# Disable OOM killer
OOMScoreAdjust=-1000
# Increase fd limit
LimitNOFILE=2048

# Systemd waits until all xlogs are recovered
TimeoutStartSec=86400s
# Give a reasonable amount of time to close xlogs
TimeoutStopSec=10s

[Install]
WantedBy=multi-user.target
Alias=new-stateboard-name
`
    expContentByFilename := map[string]string{
		fmt.Sprintf("%s.service", newAppName): expUnitContent,
		fmt.Sprintf("%s@.service", newAppName): expInstUnitContent,
		fmt.Sprintf("%s.service", newStateboardAppName): expStateboardUnitContent,
	}

	// create tmp directory
	tmpDir, err := ioutil.TempDir("", "tmp")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	err = initSystemdDir(tmpDir, &ctx)
	assert.Nil(err)

	// read systemd directory content
	systemdDirPath := filepath.Join(tmpDir, "/etc/systemd/system/")

	for filename, expContent := range expContentByFilename {
		content, err := ioutil.ReadFile(filepath.Join(systemdDirPath, filename))
		if err != nil {
			log.Fatal(err)
		}
		assert.Equal(string(content), expContent)
	}
}

func TestCheckBadSystemUnitParamsPath(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)

	var err error
	var ctx context.Ctx

	// non existing file
	ctx.Pack.SystemdUnitParamsPath = "bad-file-path"

	// create tmp directory
	tmpDir, err := ioutil.TempDir("", "tmp")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	err = initSystemdDir(tmpDir, &ctx)
	assert.EqualError(err, "Specified file with system unit params bad-file-path doesn't exists")
}
