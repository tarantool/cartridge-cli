#!/bin/bash

exec 2>&1
set -x -e

sudo systemctl stop myapp@instance_1
sudo systemctl stop myapp@instance_2

sudo systemctl daemon-reload

sudo yum remove -y myapp || true
sudo dpkg --remove myapp || true

rm -rf /tmp/e2e
