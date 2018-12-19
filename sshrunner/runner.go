package sshrunner

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/mikkeloscar/sshconfig"
	"golang.org/x/crypto/ssh"
)

// Runner executes commands via SSH
type Runner struct {
	client *ssh.Client
}

// New creates a new Runner
func (runner *Runner) New(host sshconfig.SSHHost) (*Runner, error) {
	privateKey, err := ioutil.ReadFile(host.IdentityFile)

	if err != nil {
		return nil, err
	}

	key, err := ssh.ParsePrivateKey(privateKey)

	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User:            host.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host.HostName, host.Port), config)

	if err != nil {
		return nil, err
	}

	return &Runner{
		client: client,
	}, nil
}

// Run executes the given command via SSH, with args interpolated.
func (runner *Runner) Run(command string, args ...interface{}) (string, string, error) {
	session, err := runner.client.NewSession()

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
