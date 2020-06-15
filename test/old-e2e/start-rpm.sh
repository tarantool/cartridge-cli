#!/bin/bash

exec 2>&1
set -x -e


APPNAME=${APPNAME:-myapp}
VERSION=${VERSION:-1.2.3-4}

RPM_PATH=/tmp/${APPNAME}-${VERSION}.rpm

[ -f ${RPM_PATH} ] && (sudo yum -y install ${RPM_PATH} || sudo apt-get install ${RPM_PATH})
[ -d /etc/tarantool/conf.d/ ]
sudo tee /etc/tarantool/conf.d/${APPNAME}.yml > /dev/null <<'CONFIG'
myapp.instance_1:
    alias: i1
    advertise_uri: localhost:3301
myapp.instance_2:
    alias: i2
    advertise_uri: localhost:3302
myapp-stateboard:
    listen: localhost:3310
    password: passwd
CONFIG

sudo systemctl daemon-reload

sudo systemctl start ${APPNAME}@instance_1
sudo systemctl enable ${APPNAME}@instance_1

sudo systemctl start ${APPNAME}@instance_2
sudo systemctl enable ${APPNAME}@instance_2

sudo systemctl start ${APPNAME}-stateboard
sudo systemctl enable ${APPNAME}-stateboard

sleep 1

sudo systemctl status ${APPNAME}@instance_1
sudo systemctl status ${APPNAME}@instance_2
sudo systemctl status ${APPNAME}-stateboard
