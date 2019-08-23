#!/bin/bash

exec 2>&1
set -x -e

pushd $(mktemp -d)

tarantoolctl rocks make cartridge-cli-scm-1.rockspec --chdir=/vagrant
.rocks/bin/cartridge create --name myapp

pushd ./myapp
# Here goes a bunch of tamporary hacks.
# It's because some modules aren't published to tarantool/rocks yet.
sed -e "s/'cartridge == .\+'/'cartridge == scm-1'/g" \
    -i myapp-scm-1.rockspec
cat > .cartridge.pre <<SCRIPT
#!/bin/bash -x -e
tarantoolctl rocks install https://raw.githubusercontent.com/rosik/frontend-core/gh-pages/frontend-core-5.0.2-1.rockspec
tarantoolctl rocks install https://raw.githubusercontent.com/rosik/cartridge/pre-release/cartridge-scm-1.rockspec
SCRIPT
git add .cartridge.pre
git commit -m "Add submodule"
popd

.rocks/bin/cartridge pack rpm myapp
sudo yum -y remove myapp || true
# rpm -qpl ./myapp-*.rpm
[ -f ./myapp-*.rpm ] && sudo yum -y install ./myapp-*.rpm
# sudo systemctl start myapp@i.1
# sudo systemctl start myapp@i.2
sudo yum -y remove myapp

rm -rf $(pwd)
popd
