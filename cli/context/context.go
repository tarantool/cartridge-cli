package context

import (
	"net/http"
	"time"

	"github.com/tarantool/cartridge-cli/cli/common"
)

type Ctx struct {
	Project     ProjectCtx
	Create      CreateCtx
	Build       BuildCtx
	Running     RunningCtx
	Pack        PackCtx
	Tarantool   TarantoolCtx
	Cli         CliCtx
	Docker      DockerCtx
	Repair      RepairCtx
	Admin       AdminCtx
	Replicasets ReplicasetsCtx
	Connect     ConnectCtx
	Failover    FailoverCtx
	Bench       BenchCtx
}

type ProjectCtx struct {
	Name           string
	NameToLower    string
	StateboardName string
	Path           string
}

type CreateCtx struct {
	TemplateFS http.FileSystem
	Template   string
	From       string
}

type RepairCtx struct {
	DryRun bool
	Force  bool
	Reload bool

	SetURIInstanceUUID string
	NewURI             string

	RemoveInstanceUUID string

	SetLeaderReplicasetUUID string
	SetLeaderInstanceUUID   string
}

type BuildCtx struct {
	ID   string
	Dir  string
	Spec string

	InDocker   bool
	DockerFrom string

	SDKLocal        bool
	SDKPath         string
	BuildSDKDirname string
}

type RunningCtx struct {
	Instances           []string
	WithStateboard      bool
	StateboardFlagIsSet bool
	StateboardOnly      bool

	Daemonize    bool
	StartTimeout time.Duration

	LogFollow        bool
	LogLines         int
	DisableLogPrefix bool

	StopForced bool

	Entrypoint           string
	StateboardEntrypoint string
	AppsDir              string
	AppDir               string
	ConfPath             string
	RunDir               string
	DataDir              string
	LogDir               string
}

type PackCtx struct {
	ID string

	Type string

	DockerFrom string
	NoCache    bool

	PackageFilesDir string
	ResPackagePath  string
	ResImageTags    []string

	Version           string
	Filename          string
	Release           string
	Arch              string
	Suffix            string
	VersionWithSuffix string
	ImageTags         []string

	UnitTemplatePath          string
	InstUnitTemplatePath      string
	StatboardUnitTemplatePath string

	Deps common.PackDependencies

	PreInstallScript  string
	PostInstallScript string

	PreInstallScriptFile  string
	PostInstallScriptFile string

	SystemdUnitParamsPath string
}

type TarantoolCtx struct {
	TarantoolDir           string
	TarantoolVersion       string
	TarantoolIsEnterprise  bool
	IsUserSpecifiedVersion bool
}

type CliCtx struct {
	Verbose bool
	Debug   bool
	Quiet   bool

	CartridgeTmpDir string
	TmpDir          string
	CacheDir        string
}

type DockerCtx struct {
	CacheFrom []string
}

type AdminCtx struct {
	Help bool
	List bool

	InstanceName string
	ConnString   string
}

type ReplicasetsCtx struct {
	File            string
	BootstrapVshard bool

	ReplicasetName string

	JoinInstancesNames    []string
	RolesList             []string
	VshardGroup           string
	FailoverPriorityNames []string
}

type ConnectCtx struct {
	Username string
	Password string
}

type FailoverCtx struct {
	File          string
	Mode          string
	StateProvider string

	ParamsJSON         string
	ProviderParamsJSON string
}

type BenchCtx struct {
	URL                  string // URL - the URL of the tarantool used for testing
	User                 string // User - username to connect to the tarantool.
	Password             string // Password to connect to the tarantool.
	Connections          int    // Connections describes the number of connection to be used in the test.
	SimultaneousRequests int    // SimultaneousRequests describes the number of parallel requests from one connection.
	Duration             int    // Duration describes test duration in seconds.
	KeySize              int    // DataSize describes the size of key part of benchmark data (bytes).
	DataSize             int    // DataSize describes the size of value part of benchmark data (bytes).
	InsertCount          int    // InsertCount describes the number of insert operations as a percentage.
	SelectCount          int    // SelectCount describes the number of select operations as a percentage.
	UpdateCount          int    // UpdateCount describes the number of update operations as a percentage.
	PreFillingCount      int    // PreFillingCount describes the number of records to pre-fill the space.
}
