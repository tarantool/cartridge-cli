#!/bin/bash

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
DROP='\033[0m'

APPNAME=myapp
VERSION=1.2.3-4
DEB=${APPNAME}-${VERSION}.deb

# possible values: opensource-1.10, opensource-2.3, enterprise
tarantool_version=${TARANTOOL_VERSION:-enterprise}
tarantool_repo_version=1_10

if [[ ${tarantool_version} = "enterprise" ]]; then
    if [[ `tarantool -v` != *"Tarantool Enterprise"* ]]; then
        echo -e "${RED}ERROR: You are using Tarantool Opensource${DROP}"
        exit 1;
    fi;

    if [[ -z ${TARANTOOL_SDK_PATH} ]]; then
        echo -e "${RED}ERROR: Please set TARANTOOL_SDK_PATH variable${DROP}"
        exit 1;
    fi;
fi

if [[ ${tarantool_version} = "opensource"* ]]; then
    if [[ `tarantool -v` == *"Tarantool Enterprise"* ]]; then
        echo -e "${RED}ERROR: You are using Tarantool Enterprise${DROP}"
        exit 1;
    fi;

    if [[ ${tarantool_version} = " opensource-2.3"  ]]; then
        tarantool_repo_version=2_3
    fi;

    echo "${RED}ERROR: Tests for Opesnource arent'implemented yet${DROP}";
    exit 1;
fi

rm -rf ${APPNAME}*
./cartridge create --name ${APPNAME}

# test RPM
echo -e "${GREEN}=========================================================${DROP}"
echo -e "${GREEN} Test RPM ${DROP}"
echo -e "${GREEN}---------------------------------------------------------${DROP}";

echo -e "${GREEN}---------------------------------------------------------${DROP}";
echo -e "${GREEN} Create RPM package ${DROP}"
echo -e "${GREEN}---------------------------------------------------------${DROP}";
./cartridge pack rpm --use-docker --version ${VERSION} ${APPNAME}
RPM=${APPNAME}-${VERSION}.rpm

for vm in centos alt
do
    echo -e "${GREEN}=========================================================${DROP}"
    echo -e "${GREEN} Test ${vm} ${DROP}";
    echo -e "${GREEN}---------------------------------------------------------${DROP}";
    vagrant up ${vm};

    echo -e "${GREEN}---------------------------------------------------------${DROP}";
    echo -e "${GREEN} Cleanup ${DROP}";
    echo -e "${GREEN}---------------------------------------------------------${DROP}";
    vagrant ssh ${vm} < test/e2e/cleanup.sh;

    echo -e "${GREEN}---------------------------------------------------------${DROP}";
    echo -e "${GREEN} Start RPM ${DROP}";
    echo -e "${GREEN}---------------------------------------------------------${DROP}";
    vagrant scp ${RPM} ${vm}:/tmp;
    vagrant ssh ${vm} < test/e2e/start-rpm.sh;

    echo -e "${GREEN}---------------------------------------------------------${DROP}";
    echo -e "${GREEN} Test cluster ";
    echo -e "${GREEN}---------------------------------------------------------${DROP}";
    vagrant ssh ${vm} < test/e2e/test-cluster.sh;

    echo -e "${GREEN}---------------------------------------------------------${DROP}";
    echo -e "${GREEN} Cleanup ${DROP}";
    echo -e "${GREEN}---------------------------------------------------------${DROP}";
    vagrant ssh ${vm} < test/e2e/cleanup.sh;

    vagrant halt ${vm};
    echo -e "${GREEN}=========================================================${DROP}"
done


# test DEB
echo -e "${GREEN}=========================================================${DROP}"
echo -e "${GREEN} Test DEB ${DROP}"
echo -e "${GREEN}---------------------------------------------------------${DROP}";

echo -e "${GREEN}---------------------------------------------------------${DROP}";
echo -e "${GREEN} Create DEB package ${DROP}"
echo -e "${GREEN}---------------------------------------------------------${DROP}";
./cartridge pack deb --use-docker --version ${VERSION} ${APPNAME}
DEB=${APPNAME}-${VERSION}.deb

# APPNAME=myapp ./test/e2e/cleanup.sh

for vm in ubuntu
do
    echo -e "${GREEN}=========================================================${DROP}"
    echo -e "${GREEN} Test ${vm} ${DROP}";
    echo -e "${GREEN}---------------------------------------------------------${DROP}";
    vagrant up ${vm};

    echo -e "${GREEN}---------------------------------------------------------${DROP}";
    echo -e "${GREEN} Cleanup ${DROP}";
    echo -e "${GREEN}---------------------------------------------------------${DROP}";
    vagrant ssh ${vm} < test/e2e/cleanup.sh;

    echo -e "${GREEN}---------------------------------------------------------${DROP}";
    echo -e "${GREEN} Start DEB ${DROP}";
    echo -e "${GREEN}---------------------------------------------------------${DROP}";
    vagrant scp ${DEB} ${vm}:/tmp;
    vagrant ssh ${vm} < test/e2e/start-deb.sh;

    echo -e "${GREEN}---------------------------------------------------------${DROP}";
    echo -e "${GREEN} Test cluster ";
    echo -e "${GREEN}---------------------------------------------------------${DROP}";
    vagrant ssh ${vm} < test/e2e/test-cluster.sh;

    echo -e "${GREEN}---------------------------------------------------------${DROP}";
    echo -e "${GREEN} Cleanup ${DROP}";
    echo -e "${GREEN}---------------------------------------------------------${DROP}";
    vagrant ssh ${vm} < test/e2e/cleanup.sh;

    vagrant halt ${vm};
    echo -e "${GREEN}=========================================================${DROP}"
done
