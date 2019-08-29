#!/bin/bash

mkdir -p ./dev
export HOSTNAME='localhost'
export TARANTOOL_CFG="cartridge-start-config.yml"
tarantool init.lua --cfg cartridge-start-config.yml --app-name "myapp" --instance-name "instance.1" & echo $! >> ./dev/pids
tarantool init.lua --cfg cartridge-start-config.yml --app-name "myapp" --instance-name "instance.2" & echo $! >> ./dev/pids
tarantool init.lua --cfg cartridge-start-config.yml --app-name "myapp" --instance-name "instance.3" & echo $! >> ./dev/pids
sleep 2.5
echo "All instances started!"
