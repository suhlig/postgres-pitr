package vagrant

import (
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/mikkeloscar/sshconfig"
)

// Hosts provides the SSH configuration of Vagrant VMs, indexed by the host name
func Hosts() (map[string]*sshconfig.SSHHost, error) {
	result := map[string]*sshconfig.SSHHost{}

	stdout, _, err := run("vagrant", "ssh-config")

	if err != nil {
		return result, err
	}

	tmpfile, err := ioutil.TempFile("", "vagrant-ssh-config")

	if err != nil {
		return result, err
	}

	configFileName := tmpfile.Name()
	defer os.Remove(configFileName)

	_, err = tmpfile.Write([]byte(stdout))

	if err != nil {
		return result, err
	}

	err = tmpfile.Close()

	if err != nil {
		return result, err
	}

	hosts, err := sshconfig.ParseSSHConfig(configFileName)

	if err != nil {
		return result, err
	}

	for _, host := range hosts {
		for _, hostName := range host.Host {
			result[hostName] = host
		}
	}

	return result, nil
}

func run(args ...string) (string, string, error) {
	cmd := exec.Command(args[0], args[1:]...)

	var stderr string

	out, err := cmd.Output()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			stderr = string(exitError.Stderr)
		}
	}

	return string(out), stderr, err
}
