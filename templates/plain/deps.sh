#!/bin/sh
# Call this scripts to install dependencies

set -e

tarantoolctl rocks make
tarantoolctl rocks install luacheck 0.25.0
