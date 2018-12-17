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
	Host sshconfig.SSHHost
}

// New creates a new Runner
func (runner *Runner) New(host sshconfig.SSHHost) (*Runner, error) {
	v := &Runner{Host: host}
	return v, nil
}

// Run executes the given command via SSH, with args interpolated.
func (runner *Runner) Run(command string, args ...interface{}) (string, string, error) {
	privateKey, err := ioutil.ReadFile(runner.Host.IdentityFile)

	if err != nil {
		return "", "", err
	}

	key, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return "", "", err
	}

	config := &ssh.ClientConfig{
		User:            runner.Host.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}

	addr := fmt.Sprintf("%s:%d", runner.Host.HostName, runner.Host.Port)
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
