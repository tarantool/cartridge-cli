#!/bin/bash

exec 2>&1
set -x -e

pushd $(mktemp -d)

tarantoolctl rocks make tarantoolapp-scm-1.rockspec --chdir=/vagrant
.rocks/bin/tarantoolapp create --name myapp --template cluster

pushd ./myapp
# Here goes a bunch of tamporary hacks.
# It's because some modules aren't published to tarantool/rocks yet.
sed -e "s/'cluster == .\+'/'cluster == scm-1'/g" \
    -i myapp-scm-1.rockspec
cat > .tarantoolapp.pre <<SCRIPT
#!/bin/bash -x -e
tarantoolctl rocks install https://raw.githubusercontent.com/tarantool/membership/gh-pages/membership-2.1.3-1.rockspec
tarantoolctl rocks install https://raw.githubusercontent.com/tarantool/errors/gh-pages/errors-2.1.1-1.rockspec
tarantoolctl rocks install https://raw.githubusercontent.com/rosik/frontend-core/gh-pages/frontend-core-5.0.2-1.rockspec
tarantoolctl rocks install https://raw.githubusercontent.com/rosik/cartridge/pre-release/cluster-scm-1.rockspec
SCRIPT
git add .tarantoolapp.pre
git commit -m "Add submodule"
popd

.rocks/bin/tarantoolapp pack rpm myapp
# rpm -qpl ./myapp-*.rpm
[ -f ./myapp-*.rpm ] && sudo yum -y install ./myapp-*.rpm
# sudo systemctl start myapp@i.1
# sudo systemctl start myapp@i.2
sudo yum -y remove myapp

rm -rf $(pwd)
popd
