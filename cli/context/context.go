package context

import "time"

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
}

type ProjectCtx struct {
	Name           string
	StateboardName string
	Path           string
}

type CreateCtx struct {
	Template string
	From     string
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
	ID  string
	Dir string

	InDocker   bool
	DockerFrom string

	SDKLocal        bool
	SDKPath         string
	BuildSDKDirname string
}

type RunningCtx struct {
	Instances      []string
	WithStateboard bool
	StateboardOnly bool

	Daemonize    bool
	StartTimeout time.Duration

	LogFollow bool
	LogLines  int

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

	PackageFilesDir string
	ResPackagePath  string
	ResImageTags    []string

	Version        string
	Release        string
	VersionRelease string
	Suffix         string
	ImageTags      []string

	UnitTemplatePath          string
	InstUnitTemplatePath      string
	StatboardUnitTemplatePath string
}

type TarantoolCtx struct {
	TarantoolDir          string
	TarantoolVersion      string
	TarantoolIsEnterprise bool
}

type CliCtx struct {
	Verbose bool
	Debug   bool
	Quiet   bool

	CartridgeTmpDir string
	TmpDir          string
}

type DockerCtx struct {
	NoCache   bool
	CacheFrom []string
}

type AdminCtx struct {
	Help bool
	List bool

	InstanceName string
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
