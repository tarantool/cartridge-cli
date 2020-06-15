#!/bin/bash

exec 2>&1
set -x -e

IP="localhost"
curl -w "\n" -X POST http://127.0.0.1:8081/admin/api --fail -d@- <<QUERY
{"query":
    "mutation {
        j1: join_server(
            uri:\"$IP:3301\",
            roles: [\"vshard-router\", \"app.roles.custom\", \"failover-coordinator\"]
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
        enable_failover: cluster{
            failover_params(
                mode: \"stateful\",
                tarantool_params: {
                    uri: \"$IP:3310\", password: \"passwd\"
                },
                state_provider: \"tarantool\"
            ) {
                mode
                state_provider
            }
        }
    }"
}
QUERY

sudo bash -c "export PATH=${PATH}:/usr/share/tarantool/myapp &&
tarantoolctl connect /var/run/tarantool/myapp.instance_1.control <<COMMAND
    log = require('log')
    yaml = require('yaml')
    cartridge = require('cartridge')
    cartridge_admin = require('cartridge.admin')

    assert(cartridge.is_healthy(), 'Healthcheck failed')

    s1 = cartridge_admin.get_servers('aaaaaaaa-aaaa-4000-b000-000000000001')[1]
    s2 = cartridge_admin.get_servers('bbbbbbbb-bbbb-4000-b000-000000000001')[1]
    log.info('%s', yaml.encode({s1, s2}))

    assert(s1.alias == 'i1', 'Invalid i1 alias')
    assert(s2.alias == 'i2', 'Invalid i2 alias')

    assert(s1.replicaset.roles[1] == 'failover-coordinator', 'Missing s1 failover coordinator role')
    assert(s1.replicaset.roles[2] == 'vshard-router', 'Missing s1 router role')
    assert(s1.replicaset.roles[3] == 'app.roles.custom', 'Missing s1 custom role')

    assert(s2.replicaset.roles[1] == 'vshard-storage', 'Missing s2 storage role')

    local failover_params = cartridge.failover_get_params()
    log.info('%s', yaml.encode(failover_params))

    assert(failover_params.mode == 'stateful', 'Failover mode should be stateful')
    assert(failover_params.state_provider == 'tarantool', 'Failover state provider should be tarantool')
COMMAND"
echo " - Cluster is ready"
