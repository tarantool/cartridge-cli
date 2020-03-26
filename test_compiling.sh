#!/usr/bin/env bash

set -xe

TEST_DIRNAME=test-compile
CLI_DIR=$(pwd)
APPNAME=myapp

rm -rf ${TEST_DIRNAME} || true
mkdir ${TEST_DIRNAME}
pushd ${TEST_DIRNAME}

tarantoolctl rocks make --chdir ${CLI_DIR}

CLI=$(pwd)/.rocks/bin/cartridge

${CLI} --version
${CLI} create --name ${APPNAME}

pushd ${APPNAME}
tree .

${CLI} build

${CLI} start -d
sleep 3
${CLI} stop

${CLI} pack tgz

popd  # APPNAME
popd  # TEST_DIRNAME
