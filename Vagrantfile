require_relative 'lib/pitr/config'
db = PITR::Config::DB.new('config.yml')

Vagrant.configure('2') do |config|
  config.vm.box = 'ubuntu/bionic64'
  config.vm.network 'forwarded_port', guest: 5432, host: db.port

  config.vm.provision 'ansible' do |ansible|
    ansible.playbook = 'ansible/playbook.yml'
    ansible.compatibility_mode = '2.0'
  end
end
