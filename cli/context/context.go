package context

import "time"

type Ctx struct {
	Project   ProjectCtx
	Build     BuildCtx
	Running   RunningCtx
	Pack      PackCtx
	Tarantool TarantoolCtx
	Cli       CliCtx
	Docker    DockerCtx
}

type ProjectCtx struct {
	Name           string
	StateboardName string
	Path           string
	Template       string
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
