package cluster

import (
	pitr "github.com/suhlig/postgres-pitr"
	"golang.org/x/crypto/ssh"
)

// Controller provides a way to control a PostgreSQL cluster with the given version and name.
// It performs its actions using the provided Runner.
type Controller struct {
	runner  pitr.Runner
	version string
	name    string
}

// NewController creates a new controller for the cluster with the given version and name
// All actions will be performed using the passed Runner.
func NewController(runner pitr.Runner, version, name string) Controller {
	controller := Controller{}
	controller.runner = runner
	controller.version = version
	controller.name = name

	return controller
}

// Start starts the cluster
func (ctl Controller) Start() *pitr.Error {
	stdout, stderr, err := ctl.runner.Run("sudo pg_ctlcluster %s %s start", ctl.version, ctl.name)

	if err != nil {
		return &pitr.Error{"Could not start the cluster", stdout, stderr}
	}

	return nil
}

// IsRunning returns true if the cluster is running
func (ctl Controller) IsRunning() (bool, *pitr.Error) {
	stdout, stderr, err := ctl.runner.Run("sudo pg_ctlcluster %s %s status", ctl.version, ctl.name)

	if result, ok := err.(*ssh.ExitError); ok {
		if result.ExitStatus() == 3 { // server is stopped
			return false, nil
		}

		// something else is going on
		return false, &pitr.Error{
			Message: err.Error(),
			Stdout:  stdout,
			Stderr:  stderr,
		}
	}

	// server is running
	return true, nil
}

// Stop stops the cluster, if running.
func (ctl Controller) Stop() *pitr.Error {
	running, err := ctl.IsRunning()

	if err != nil {
		return err
	}

	if !running {
		return nil
	}

	stdout, stderr, runErr := ctl.runner.Run("sudo pg_ctlcluster %s %s stop", ctl.version, ctl.name)

	if runErr != nil {
		return &pitr.Error{
			Message: "Could not stop the cluster",
			Stdout:  stdout,
			Stderr:  stderr,
		}
	}

	return nil
}
