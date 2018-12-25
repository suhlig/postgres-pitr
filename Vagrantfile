require_relative 'lib/pitr/config'
require 'pathname'

db = PITR::Config::DB.new(Pathname(__dir__) / 'config.yml')

Vagrant.configure('2') do |config|
  config.vm.box = 'ubuntu/bionic64'

  config.vm.define 'postgres' do |cfg|
    cfg.vm.hostname = 'postgres'
    cfg.vm.network 'private_network', ip: db.host
    cfg.vm.network 'forwarded_port', guest: db.port, host: db.local_port
    cfg.vm.post_up_message = "PostgreSQL available at #{db.local_url}"
  end

  config.vm.provision 'ansible' do |ansbl|
    ansbl.playbook = 'ansible/playbook.yml'
    ansbl.compatibility_mode = '2.0'
    ansbl.extra_vars = {
      ansible_python_interpreter: '/usr/bin/python3',
    }
    ansbl.groups = {
      'databases' => ['postgres'],
    }
  end

  # https://stackoverflow.com/a/37946223
  config.vm.provider 'virtualbox' do |vb|
    vb.customize [ "guestproperty", "set", :id, "/VirtualBox/GuestAdd/VBoxService/--timesync-set-threshold", 1000 ]
  end
end
