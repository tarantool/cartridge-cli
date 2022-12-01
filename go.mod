module github.com/tarantool/cartridge-cli

go 1.17

require (
	github.com/FZambia/tarantool v0.2.1
	github.com/adam-hanna/arrayOperations v0.2.6
	github.com/alecthomas/participle/v2 v2.0.0-alpha4
	github.com/apex/log v1.4.0
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/briandowns/spinner v1.11.1
	github.com/c-bata/go-prompt v0.2.5
	github.com/dave/jennifer v1.4.1
	github.com/docker/docker v20.10.7+incompatible
	github.com/fatih/color v1.7.0
	github.com/fatih/structs v1.1.0
	github.com/hashicorp/go-version v1.2.0
	github.com/hpcloud/tail v1.0.0
	github.com/magefile/mage v1.11.0
	github.com/mattn/go-isatty v0.0.12
	github.com/mitchellh/mapstructure v1.4.1
	github.com/otiai10/copy v1.7.1
	github.com/pmezard/go-difflib v1.0.0
	github.com/robfig/config v0.0.0-20141207224736-0f78529c8c7e
	github.com/shirou/gopsutil v3.21.2+incompatible
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/spf13/cobra v1.0.1-0.20200815144417-81e0311edd0b
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/vmihailenco/msgpack/v5 v5.1.0
	github.com/yuin/gopher-lua v0.0.0-20191220021717-ab39c6098bdb
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/Microsoft/go-winio v0.4.17 // indirect
	github.com/StackExchange/wmi v0.0.0-20210224194228-fe8f1750fd46 // indirect
	github.com/containerd/containerd v1.5.13 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mattn/go-tty v0.0.3 // indirect
	github.com/moby/term v0.0.0-20221105221325-4eb28fa6025c // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/term v1.2.0-beta.2 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/tklauser/go-sysconf v0.3.4 // indirect
	github.com/tklauser/numcpus v0.2.1 // indirect
	github.com/vmihailenco/tagparser v0.1.2 // indirect
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
	golang.org/x/tools v0.1.5 // indirect
	google.golang.org/genproto v0.0.0-20201110150050-8816d57aaa9a // indirect
	google.golang.org/grpc v1.39.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/fsnotify.v1 v1.4.7 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

replace (
	github.com/c-bata/go-prompt => github.com/tarantool/go-prompt v0.2.6-tarantool
	github.com/docker/docker/internal/testutil => gotest.tools/v3 v3.0.0
	golang.org/x/crypto => golang.org/x/crypto v0.0.0-20211202192323-5770296d904e
	golang.org/x/sys => golang.org/x/sys v0.0.0-20210817190340-bfb29a6856f2
	golang.org/x/text => golang.org/x/text v0.3.7
)
