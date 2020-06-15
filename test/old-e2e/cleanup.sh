#!/bin/bash

exec 2>&1
set -x -e

APPNAME=${APPNAME:-myapp}

sudo systemctl stop ${APPNAME}@instance_1 || true
sudo systemctl stop ${APPNAME}@instance_2 || true

sudo systemctl daemon-reload

sudo yum remove -y ${APPNAME} || true
sudo dpkg --remove ${APPNAME} || true
sudo apt-get remove ${APPNAME} || true

rm -rf /tmp/e2e
