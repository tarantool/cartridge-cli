#!/bin/bash

exec 2>&1
set -x -e

git config --global user.email "you@example.com"
git config --global user.name "Your Name"

pushd $(mktemp -d)

tarantoolctl rocks make cartridge-cli-scm-1.rockspec --chdir=/vagrant
.rocks/bin/cartridge create --name myapp

sudo yum -y remove myapp || true
sudo rm -rf /etc/tarantool/conf.d || true
sudo chmod +x /var/lib/tarantool/ || true
sudo rm -rf /var/lib/tarantool/myapp.instance_{1,2} || true

.rocks/bin/cartridge pack rpm myapp
[ -f ./myapp-*.rpm ] && sudo yum -y install ./myapp-*.rpm
[ -d /etc/tarantool/conf.d/ ]
sudo tee /etc/tarantool/conf.d/myapp.yml > /dev/null <<'CONFIG'
myapp.instance_1:
    alias: i1
myapp.instance_2:
    alias: i2
CONFIG

sudo systemctl daemon-reload

sudo systemctl start myapp@instance_1
sudo systemctl enable myapp@instance_1

sudo systemctl start myapp@instance_2
sudo systemctl enable myapp@instance_2

sleep 1

popd
