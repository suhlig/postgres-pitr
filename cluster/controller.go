package cluster

import (
	"fmt"
)

type Runner interface {
	Run(command string, args ...interface{}) (string, string, error)
}

// Controller provides a way to control a PostgreSQL cluster
type Controller struct {
	Runner  Runner
	Version string
	Name    string
}

// NewController creates a new controller
func NewController(runner Runner, version, name string) (Controller, error) {
	controller := Controller{}
	controller.Runner = runner
	controller.Version = version
	controller.Name = name

	return controller, nil
}

// Start starts the cluster with the given version and name
func (ctl Controller) Start() error {
	stdout, stderr, err := ctl.Runner.Run("sudo pg_ctlcluster %s %s start", ctl.Version, ctl.Name)

	if err != nil {
		return fmt.Errorf("Error: %v\nstderr:\n%v\nstdout:\n%v\n", err, stdout, stderr)
	}

	return nil
}

// Stop stops the cluster with the given version and name
func (ctl Controller) Stop() error {
	stdout, stderr, err := ctl.Runner.Run("sudo pg_ctlcluster %s %s stop", ctl.Version, ctl.Name)

	if err != nil {
		return fmt.Errorf("Error: %v\nstderr:\n%v\nstdout:\n%v\n", err, stdout, stderr)
	}

	return nil
}

// Clear removes all files of the cluster with the given version and name
func (ctl Controller) Clear() error {
	stdout, stderr, err := ctl.Runner.Run("sudo -u postgres find /var/lib/postgresql/%s/%s -mindepth 1 -delete", ctl.Version, ctl.Name)

	if err != nil {
		return fmt.Errorf("Error: %v\nstderr:\n%v\nstdout:\n%v\n", err, stdout, stderr)
	}

	return nil
}
