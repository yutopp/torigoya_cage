# -*- mode: ruby -*-
# vi: set ft=ruby :

# Vagrantfile API/syntax version. Don't touch unless you know what you're doing!
VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
  # All Vagrant configuration is done here. The most common configuration
  # options are documented and commented below. For a complete reference,
  # please see the online documentation at vagrantup.com.

  # Every Vagrant virtual environment requires a box to build off of.
  config.vm.box = "ubuntu/trusty64"

  # port 12321 is used by torigoya cage server
  config.vm.network "forwarded_port", guest: 12321, host: 12321, auto_correct: true

  config.vm.provider :virtualbox do |vb|
    vb.customize ["modifyvm", :id, "--memory", 1024]
    # http://stackoverflow.com/questions/22901859/cannot-make-outbound-http-requests-from-vagrant-vm
    vb.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
  end

  # https://coderwall.com/p/qtbi5a
  config.ssh.shell = "bash -c 'BASH_ENV=/etc/profile exec bash'"

  #
  config.vm.provision :shell, :inline => "cp /vagrant/torigoya-packages.list /etc/apt/sources.list.d/."
end
