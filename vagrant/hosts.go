package vagrant

import (
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/mikkeloscar/sshconfig"
)

// Hosts provides the SSH configuration of Vagrant VMs
func Hosts() ([]*sshconfig.SSHHost, error) {
	stdout, _, err := run("vagrant", "ssh-config")
	if err != nil {
		return nil, err
	}

	tmpfile, err := ioutil.TempFile("", "vagrant-ssh-config")
	if err != nil {
		return nil, err
	}

	configFileName := tmpfile.Name()
	defer os.Remove(configFileName)

	_, err = tmpfile.Write([]byte(stdout))
	if err != nil {
		return nil, err
	}

	err = tmpfile.Close()
	if err != nil {
		return nil, err
	}

	return sshconfig.ParseSSHConfig(configFileName)
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
