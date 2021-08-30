module github.com/tarantool/cartridge-cli

go 1.16

require (
	github.com/FZambia/tarantool v0.2.1
	github.com/StackExchange/wmi v0.0.0-20210224194228-fe8f1750fd46 // indirect
	github.com/adam-hanna/arrayOperations v0.2.6
	github.com/alecthomas/participle/v2 v2.0.0-alpha4
	github.com/apex/log v1.4.0
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/briandowns/spinner v1.11.1
	github.com/c-bata/go-prompt v0.2.5
	github.com/containerd/containerd v1.5.4 // indirect
	github.com/dave/jennifer v1.4.1
	github.com/docker/docker v20.10.7+incompatible
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/fatih/color v1.7.0
	github.com/fatih/structs v1.1.0
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/hashicorp/go-version v1.2.0
	github.com/hpcloud/tail v1.0.0
	github.com/magefile/mage v1.11.0
	github.com/mattn/go-isatty v0.0.12
	github.com/mitchellh/mapstructure v1.4.1
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/otiai10/copy v1.2.0
	github.com/pmezard/go-difflib v1.0.0
	github.com/shirou/gopsutil v3.21.2+incompatible
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/vfsgen v0.0.0-20200824052919-0d455de96546 // indirect
	github.com/spf13/cobra v1.0.1-0.20200815144417-81e0311edd0b
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/tklauser/go-sysconf v0.3.4 // indirect
	github.com/vmihailenco/msgpack/v5 v5.1.0
	github.com/yuin/gopher-lua v0.0.0-20191220021717-ab39c6098bdb
	google.golang.org/grpc v1.39.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/docker/docker/internal/testutil => gotest.tools/v3 v3.0.0

replace github.com/c-bata/go-prompt => github.com/tarantool/go-prompt v0.2.6-tarantool

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20210817190340-bfb29a6856f2
