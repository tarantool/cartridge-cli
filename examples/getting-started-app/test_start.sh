#!/bin/bash
set -e

tarantoolctl rocks make
tarantoolctl rocks install cartridge-cli

export PATH=$(pwd)/.rocks/bin/:$PATH

./start.sh /tmp/cartridge_example_test

echo "Check instances... "
sleep 5
tarantool check_instances.lua \
    && export TNT_GSE_STATUS="Success" \
    || export TNT_GSE_STATUS=""

if [[ "$TNT_GSE_STATUS" ]]; then
    echo "\t\033[0;32mSuccess\033[0m"
else 
    echo "\t\033[0;31mFailed\033[0m"
fi

./stop.sh /tmp/cartridge_example_test

if [[ "$TNT_GSE_STATUS" ]]; then
    .rocks/bin/luatest -v 
else 
    echo "API access error"
    exit 1
fi
