#!/bin/bash

exec 2>&1
set -x -e

git config --global user.email "you@example.com"
git config --global user.name "Your Name"

pushd $(mktemp -d)

tarantoolctl rocks make cartridge-cli-scm-1.rockspec --chdir=/vagrant
.rocks/bin/cartridge create --name myapp

sudo dpkg --remove myapp || true
sudo rm -rf /etc/tarantool/conf.d || true
sudo chmod +x /var/lib/tarantool/ || true
sudo rm -rf /var/lib/tarantool/myapp* || true

.rocks/bin/cartridge pack deb myapp
[ -f ./myapp-*.deb ] && sudo dpkg -i ./myapp-*.deb
[ -d /etc/tarantool/conf.d/ ]
sudo tee /etc/tarantool/conf.d/myapp.yml > /dev/null <<'CONFIG'
myapp.instance_1:
    alias: i1
myapp.instance_2:
    alias: i2
CONFIG

sudo systemctl start myapp@instance_1
sudo systemctl start myapp@instance_2
sleep 0.5

IP=$(hostname -I | tr -d '[:space:]')
curl -w "\n" -X POST http://127.0.0.1:8081/admin/api --fail -d@- <<QUERY
{"query":
    "mutation {
        j1: join_server(
            uri:\"$IP:3301\",
            roles: [\"vshard-router\", \"app.roles.custom\"]
            instance_uuid: \"aaaaaaaa-aaaa-4000-b000-000000000001\"
            replicaset_uuid: \"aaaaaaaa-0000-4000-b000-000000000000\"
        )
        j2: join_server(
            uri:\"$IP:3302\",
            roles: [\"vshard-storage\"]
            instance_uuid: \"bbbbbbbb-bbbb-4000-b000-000000000001\"
            replicaset_uuid: \"bbbbbbbb-0000-4000-b000-000000000000\"
            timeout: 5
        )
        bootstrap_vshard
    }"
}
QUERY
sudo tarantoolctl connect /var/run/tarantool/myapp.instance_1.control <<COMMAND
    log = require('log')
    yaml = require('yaml')
    cartridge = require('cartridge')
    cartridge_admin = require('cartridge.admin')

    assert(cartridge.is_healthy(), "Healthcheck failed")

    s1 = cartridge_admin.get_servers('aaaaaaaa-aaaa-4000-b000-000000000001')[1]
    s2 = cartridge_admin.get_servers('bbbbbbbb-bbbb-4000-b000-000000000001')[1]
    log.info('%s', yaml.encode({s1, s2}))

    assert(s1.alias == 'i1', "Invalid i1 alias")
    assert(s2.alias == 'i2', "Invalid i2 alias")

    assert(s1.replicaset.roles[1] == 'vshard-router', "Missing s1 router role")
    assert(s2.replicaset.roles[1] == 'vshard-storage', "Missing s2 storage role")
    assert(s1.replicaset.roles[2] == 'app.roles.custom', "Missing s1 custom role")
COMMAND
echo " - Cluster is ready"

sudo systemctl stop myapp@instance_1
sudo systemctl stop myapp@instance_2

sudo dpkg --remove myapp

rm -rf $(pwd)
popd
