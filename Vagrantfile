# -*- mode: ruby -*-
# vi: set ft=ruby :

# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure("2") do |config|
  config.vm.define "centos" do |cfg|
    cfg.vm.box = "centos/7"
    cfg.vm.provision "shell", inline: <<-SHELL
      curl -s https://packagecloud.io/install/repositories/tarantool/1_10/script.rpm.sh | bash
      curl -sL https://rpm.nodesource.com/setup_8.x | bash -
      yum -y install unzip git gcc cmake nodejs tarantool tarantool-devel
    SHELL
  end

  config.vm.define "ubuntu" do |cfg|
    cfg.vm.box = "ubuntu/xenial64"
    cfg.vm.provision "shell", inline: <<-SHELL
      curl -s https://packagecloud.io/install/repositories/tarantool/1_10/script.deb.sh | bash
      sudo apt-get -y update
      sudo apt-get -y install unzip git gcc cmake nodejs tarantool
      sudo rm /lib/systemd/system/tarantool@.service
    SHELL
  end
end
