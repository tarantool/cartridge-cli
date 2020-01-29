#!/bin/sh
# Call this scripts to install dependencies

set -e

tarantoolctl rocks make
# Dev and test dependencies:
tarantoolctl rocks install luatest 0.5.0
tarantoolctl rocks install luacov 0.13.0
tarantoolctl rocks install luacheck 0.25.0
