package pgbackrest

import (
	"encoding/json"
	"fmt"
	"time"

	pitr "github.com/suhlig/postgres-pitr"
)

// Info tells about the backups of a stanza
type Info struct {
	Name   string
	Status struct {
		Code    int
		Message string
	}
}

// Controller provides a way to control pgbackrest
type Controller struct {
	Runner pitr.Runner
}

// NewController creates a new controller
func NewController(runner pitr.Runner) (Controller, error) {
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

// Backup creates a new backup for the given stanza
func (ctl Controller) Backup(stanza string) error {
	stdout, stderr, err := ctl.Runner.Run("sudo -u postgres pgbackrest --stanza=%s backup --type=incr", stanza)

	if err != nil {
		return fmt.Errorf("Error: %v\nstderr:\n%v\nstdout:\n%v\n", err, stdout, stderr)
	}

	return nil
}

// Restore a backup for the given stanza
func (ctl Controller) Restore(stanza string) error {
	stdout, stderr, err := ctl.Runner.Run("sudo -u postgres pgbackrest --stanza=%s restore", stanza)

	if err != nil {
		return fmt.Errorf("Error: %v\nstderr:\n%v\nstdout:\n%v\n", err, stdout, stderr)
	}

	return nil
}

// RestoreTo a specific point in time
func (ctl Controller) RestoreTo(stanza string, pointInTime time.Time) error {
	pointInTimeRFC := fmt.Sprintf((pointInTime.Format(time.RFC3339)))
	stdout, stderr, err := ctl.Runner.Run(
		"sudo -u postgres pgbackrest --stanza=%s --delta --type=time \"--target=%s\" restore",
		stanza,
		pointInTimeRFC,
	)

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
