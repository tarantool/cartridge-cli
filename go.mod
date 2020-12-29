module github.com/tarantool/cartridge-cli

go 1.16

require (
	docker.io/go-docker v1.0.0
	github.com/FZambia/tarantool v0.1.1
	github.com/Microsoft/go-winio v0.4.16 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/adam-hanna/arrayOperations v0.2.6
	github.com/apex/log v1.4.0
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/briandowns/spinner v1.11.1
	github.com/c-bata/go-prompt v0.2.5
	github.com/dave/jennifer v1.4.1
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker/internal/testutil v0.0.0-00010101000000-000000000000 // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/fatih/color v1.7.0
	github.com/fatih/structs v1.1.0
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/hashicorp/go-version v1.2.0
	github.com/hpcloud/tail v1.0.0
	github.com/magefile/mage v1.9.0
	github.com/mattn/go-isatty v0.0.12
	github.com/mitchellh/mapstructure v1.4.1
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/otiai10/copy v1.2.0
	github.com/pmezard/go-difflib v1.0.0
	github.com/shirou/gopsutil v2.20.5+incompatible
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/spf13/cobra v1.0.1-0.20200815144417-81e0311edd0b
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	github.com/vmihailenco/msgpack/v5 v5.1.0
	github.com/yuin/gopher-lua v0.0.0-20191220021717-ab39c6098bdb
	golang.org/x/tools v0.0.0-20200609124132-5359b67ffbdf // indirect
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/docker/docker/internal/testutil => gotest.tools/v3 v3.0.0
