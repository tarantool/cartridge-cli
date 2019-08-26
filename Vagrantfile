# -*- mode: ruby -*-
# vi: set ft=ruby :

# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure("2") do |config|
  config.vm.define "1_10" do |cfg|
    cfg.vm.box = "centos/7"
    cfg.vm.provision "shell", inline: <<-SHELL
      curl -s https://packagecloud.io/install/repositories/tarantool/1_10/script.rpm.sh | bash
      curl -sL https://rpm.nodesource.com/setup_8.x | bash -
      yum -y install git gcc cmake nodejs tarantool tarantool-devel
      git config --global user.email "test@tarantool.io"
      git config --global user.name "Taran Tool"
    SHELL
  end
end
