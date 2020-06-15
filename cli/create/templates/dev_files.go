package templates

import "github.com/tarantool/cartridge-cli/cli/templates"

var devFilesTemplate = templates.FileTreeTemplate{
	Dirs: []templates.DirTemplate{
		{
			Path: "tmp",
			Mode: 0755,
		},
	},
	Files: []templates.FileTemplate{
		{
			Path:    "deps.sh",
			Mode:    0755,
			Content: depsScriptContent,
		},
		{
			Path:    "instances.yml",
			Mode:    0644,
			Content: instancesConfContent,
		},
		{
			Path:    ".cartridge.yml",
			Mode:    0644,
			Content: cartridgeConfContent,
		},
		{
			Path:    "tmp/.keep",
			Mode:    0644,
			Content: "",
		},
	},
}

const (
	depsScriptContent = `#!/bin/sh
# Call this script to install test dependencies

set -e

# Test dependencies:
tarantoolctl rocks install luatest 0.5.0
tarantoolctl rocks install luacov 0.13.0
tarantoolctl rocks install luacheck 0.25.0
`

	instancesConfContent = `---
{{ .Name }}.router:
  advertise_uri: localhost:3301
  http_port: 8081
{{ .Name }}.s1-master:
  advertise_uri: localhost:3302
  http_port: 8082
{{ .Name }}.s1-replica:
  advertise_uri: localhost:3303
  http_port: 8083
{{ .Name }}.s2-master:
  advertise_uri: localhost:3304
  http_port: 8084
{{ .Name }}.s2-replica:
  advertise_uri: localhost:3305
  http_port: 8085
{{ .StateboardName }}:
  listen: localhost:3310
  password: passwd
`

	cartridgeConfContent = `---
# here you can specify "cartridge start" parametes, such as

# cfg: path-to-cfg-file
# log-dir: path-to-log-dir
# run-dir: path-to-run-dir
# data-dir: path-to-data-dir
`
)
