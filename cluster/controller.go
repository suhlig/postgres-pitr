package cluster

import "fmt"

type Controller struct {
	Runner Runner
}

type Runner interface {
	Run(command string, args ...interface{}) (string, string, error)
}

func NewController(runner Runner) (Controller, error) {
	controller := Controller{}
	controller.Runner = runner

	return controller, nil
}

func (ctl Controller) Start(version, name string) error {
	stdout, stderr, err := ctl.Runner.Run("sudo pg_ctlcluster %s %s start", version, name)

	if err != nil {
		return fmt.Errorf("Error: %v\nstderr:\n%v\nstdout:\n%v\n", err, stdout, stderr)
	}

	return nil
}

func (ctl Controller) Stop(version, name string) error {
	stdout, stderr, err := ctl.Runner.Run("sudo pg_ctlcluster %s %s stop", version, name)

	if err != nil {
		return fmt.Errorf("Error: %v\nstderr:\n%v\nstdout:\n%v\n", err, stdout, stderr)
	}

	return nil
}
