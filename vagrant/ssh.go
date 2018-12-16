package vagrant

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/mikkeloscar/sshconfig"
	"golang.org/x/crypto/ssh"
)

type VagrantSSH struct {
	Host sshconfig.SSHHost
}

func (vagrant *VagrantSSH) New(host sshconfig.SSHHost) (*VagrantSSH, error) {
	v := &VagrantSSH{Host: host}
	return v, nil
}

func (vagrant *VagrantSSH) Run(command string, args ...interface{}) (string, string, error) {
	privateKey, err := ioutil.ReadFile(vagrant.Host.IdentityFile)

	if err != nil {
		return "", "", err
	}

	key, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return "", "", err
	}

	config := &ssh.ClientConfig{
		User:            vagrant.Host.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}

	addr := fmt.Sprintf("%s:%d", vagrant.Host.HostName, vagrant.Host.Port)
	client, err := ssh.Dial("tcp", addr, config)

	if err != nil {
		return "", "", err
	}

	session, err := client.NewSession()

	if err != nil {
		return "", "", err
	}

	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	err = session.Run(fmt.Sprintf(command, args...))

	return stdoutBuf.String(), stderrBuf.String(), err
}
