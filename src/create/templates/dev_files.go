package templates

var devFilesTemplate = projectTemplate{
	Dirs: []dirTemplate{
		dirTemplate{
			Path: "tmp",
			Mode: 0755,
		},
	},
	Files: []fileTemplate{
		fileTemplate{
			Path:    "deps.sh",
			Mode:    0755,
			Content: depsScriptContent,
		},

		fileTemplate{
			Path:    "instances.yml",
			Mode:    0644,
			Content: instancesConfContent,
		},

		fileTemplate{
			Path:    ".cartridge.yml",
			Mode:    0644,
			Content: cartridgeConfContent,
		},

		fileTemplate{
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
  workdir: ./tmp/db_dev/3301
  advertise_uri: localhost:3301
  http_port: 8081
{{ .Name }}.s1-master:
  workdir: ./tmp/db_dev/3302
  advertise_uri: localhost:3302
  http_port: 8082
{{ .Name }}.s1-replica:
  workdir: ./tmp/db_dev/3303
  advertise_uri: localhost:3303
  http_port: 8083
{{ .Name }}.s2-master:
  workdir: ./tmp/db_dev/3304
  advertise_uri: localhost:3304
  http_port: 8084
{{ .Name }}.s2-replica:
  workdir: ./tmp/db_dev/3305
  advertise_uri: localhost:3305
  http_port: 8085
{{ .StateboardName }}:
  workdir: ./tmp/db_dev/3310
  listen: localhost:3310
  password: passwd
`

	cartridgeConfContent = `---
run_dir: tmp
`
)
