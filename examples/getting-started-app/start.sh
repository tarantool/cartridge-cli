#!/bin/bash

mkdir -p ./dev
export TARANTOOL_CFG=demo.yml
tarantool init.lua --instance-name "router"     & echo $! >> ./dev/pids
tarantool init.lua --instance-name "s1-master"  & echo $! >> ./dev/pids
tarantool init.lua --instance-name "s1-replica" & echo $! >> ./dev/pids
tarantool init.lua --instance-name "s2-master"  & echo $! >> ./dev/pids
tarantool init.lua --instance-name "s2-replica" & echo $! >> ./dev/pids
sleep 2.5
echo "All instances started!"
