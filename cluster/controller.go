package cluster

import (
	"fmt"

	pitr "github.com/suhlig/postgres-pitr"
	"golang.org/x/crypto/ssh"
)

// Controller provides a way to control a PostgreSQL cluster with the given version and name.
// It performs its actions using the provided Runner.
type Controller struct {
	Runner  pitr.Runner
	Version string
	Name    string
}

// Error encapsulates information about a Controller error
type Error struct {
	message        string
	stdout, stderr string
}

// NewController creates a new controller for the cluster with the given version and name
// All actions will be performed using the passed Runner.
func NewController(runner pitr.Runner, version, name string) (Controller, error) {
	controller := Controller{}
	controller.Runner = runner
	controller.Version = version
	controller.Name = name

	return controller, nil
}

// Start starts the cluster
func (ctl Controller) Start() *Error {
	stdout, stderr, err := ctl.Runner.Run("sudo pg_ctlcluster %s %s start", ctl.Version, ctl.Name)

	if err != nil {
		return &Error{"Could not start the cluster", stdout, stderr}
	}

	return nil
}

// IsRunning returns true if the cluster is running
func (ctl Controller) IsRunning() (bool, error) {
	_, _, err := ctl.Runner.Run("sudo pg_ctlcluster %s %s status", ctl.Version, ctl.Name)

	if result, ok := err.(*ssh.ExitError); ok {
		if result.ExitStatus() == 3 {
			// server is stopped
			return false, nil
		}

		// something else is going on
		return false, err
	}

	// server is running
	return true, nil
}

// Stop stops the cluster, if running.
func (ctl Controller) Stop() error {
	running, err := ctl.IsRunning()

	if err != nil {
		return err
	}

	if !running {
		return nil
	}

	stdout, stderr, err := ctl.Runner.Run("sudo pg_ctlcluster %s %s stop", ctl.Version, ctl.Name)

	if err != nil {
		return &Error{"Could not stop the cluster", stdout, stderr}
	}

	return nil
}

func (e *Error) Error() string {
	return fmt.Sprintf("Error: %s\nstderr: %s\nstdout: %s\n", e.message, e.stdout, e.stderr)
}
