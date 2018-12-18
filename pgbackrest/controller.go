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

// Controller provides a way to control pgbackrest
type Controller struct {
	Runner Runner
}

// Runner executes commands via SSH
type Runner interface {
	Run(command string, args ...interface{}) (string, string, error)
}

// NewController creates a new controller
func NewController(runner Runner) (Controller, error) {
	controller := Controller{}
	controller.Runner = runner

	return controller, nil
}

// Info provides a summary of backups for the given stanza
func (ctl Controller) Info(stanza string) ([]Info, error) {
	stdout, stderr, err := ctl.Runner.Run("sudo -u postgres pgbackrest info --stanza=%s --output=json", stanza)

	if err != nil {
		return nil, fmt.Errorf("Error: %v\nstderr:\n%v\nstdout:\n%v\n", err, stdout, stderr)
	}

	return parseInfo(stdout)
}

// Creates a new backup for the given stanza
func (ctl Controller) Backup(stanza string) error {
	stdout, stderr, err := ctl.Runner.Run("sudo -u postgres pgbackrest --stanza=%s backup --type=incr", stanza)

	if err != nil {
		return fmt.Errorf("Error: %v\nstderr:\n%v\nstdout:\n%v\n", err, stdout, stderr)
	}

	return nil
}

// Restores a backup for the given stanza
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
