require_relative 'lib/pitr/config'
require 'pathname'

master = PITR::Config::DB.new(Pathname(__dir__) / 'config.yml', 'master')
standby = PITR::Config::DB.new(Pathname(__dir__) / 'config.yml', 'standby')
minio = PITR::Config::Blobstore.new(Pathname(__dir__) / 'config.yml', 'minio')

Vagrant.configure('2') do |config|
  config.vm.box = 'ubuntu/bionic64'

  config.vm.define 'minio' do |cfg|
    cfg.vm.hostname = 'minio'
    cfg.vm.network 'private_network', ip: minio.host
    cfg.vm.network 'forwarded_port', guest: minio.port, host: minio.local_port

    # TODO Create minio.local_url
    cfg.vm.post_up_message = "Minio can be browsed at https://localhost:#{minio.local_port}/"
  end

  config.vm.define 'master' do |cfg|
    cfg.vm.hostname = 'master'
    cfg.vm.network 'private_network', ip: master.host
    cfg.vm.network 'forwarded_port', guest: master.port, host: master.local_port
    cfg.vm.post_up_message = "PostgreSQL available at #{master.local_url}"
  end

  config.vm.define 'standby' do |cfg|
    cfg.vm.hostname = 'standby'
    cfg.vm.network 'private_network', ip: standby.host
    cfg.vm.network 'forwarded_port', guest: standby.port, host: standby.local_port
    cfg.vm.post_up_message = "PostgreSQL hot standby available at #{standby.local_url}"
  end

  config.vm.provision 'ansible' do |ansbl|
    ansbl.playbook = 'ansible/playbook.yml'
    ansbl.compatibility_mode = '2.0'
    ansbl.extra_vars = {
      ansible_python_interpreter: '/usr/bin/python3',
    }
    ansbl.groups = {
      'db-masters' => ['master'],
      'db-standbys' => ['standby'],
      'blobstores' => ['minio'],
    }
  end

  # https://stackoverflow.com/a/37946223
  config.vm.provider 'virtualbox' do |vb|
    vb.customize [ "guestproperty", "set", :id, "/VirtualBox/GuestAdd/VBoxService/--timesync-set-threshold", 1000 ]
  end
end
