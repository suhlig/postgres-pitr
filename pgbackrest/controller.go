package pgbackrest

import "fmt"
import "encoding/json"

type Info struct {
	Name   string
	Status struct {
		Code    int
		Message string
	}
}

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

func (ctl Controller) Info(stanza string) ([]Info, error) {
	stdout, stderr, err := ctl.Runner.Run("sudo -u postgres pgbackrest info --stanza=%s --output=json", stanza)

	if err != nil {
		return nil, fmt.Errorf("Error: %v\nstderr:\n%v\nstdout:\n%v\n", err, stdout, stderr)
	}

	return parseInfo(stdout)
}

func (ctl Controller) Backup(stanza string) error {
	stdout, stderr, err := ctl.Runner.Run("sudo -u postgres pgbackrest --stanza=%s backup", stanza)

	if err != nil {
		return fmt.Errorf("Error: %v\nstderr:\n%v\nstdout:\n%v\n", err, stdout, stderr)
	}

	return nil
}

func (ctl Controller) Restore(stanza string) error {
	stdout, stderr, err := ctl.Runner.Run("sudo -u postgres pgbackrest --stanza=%s restore", stanza)

	if err != nil {
		return fmt.Errorf("Error: %v\nstderr:\n%v\nstdout:\n%v\n", err, stdout, stderr)
	}

	return nil
}

func parseInfo(stdout string) ([]Info, error) {
	infos := make([]Info, 0)
	err := json.Unmarshal([]byte(stdout), &infos)
	return infos, err
}
