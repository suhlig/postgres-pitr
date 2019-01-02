package pgbackrest

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/suhlig/postgres-pitr/cluster"

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
	runner  pitr.Runner
	cluster cluster.Controller
}

// NewController creates a new controller
func NewController(runner pitr.Runner, cluster cluster.Controller) Controller {
	controller := Controller{}
	controller.runner = runner
	controller.cluster = cluster

	return controller
}

// Info provides a summary of backups for the given stanza
func (ctl Controller) Info(stanza string) ([]Info, error) {
	stdout, stderr, err := ctl.runner.Run("sudo -u postgres pgbackrest info --stanza=%s --output=json", stanza)

	if err != nil {
		return nil, fmt.Errorf("Error: %v\nstderr:\n%v\nstdout:\n%v\n", err, stdout, stderr)
	}

	return parseInfo(stdout)
}

// Backup creates a new backup for the given stanza
func (ctl Controller) Backup(stanza string) error {
	stdout, stderr, err := ctl.runner.Run("sudo -u postgres pgbackrest --stanza=%s backup --type=incr", stanza)

	if err != nil {
		return fmt.Errorf("Error: %v\nstderr:\n%v\nstdout:\n%v\n", err, stdout, stderr)
	}

	return nil
}

// Restore a backup for the given stanza
func (ctl Controller) Restore(stanza string) error {
	err := ctl.cluster.Stop()

	if err != nil {
		return err
	}

	stdout, stderr, err := ctl.runner.Run("sudo -u postgres pgbackrest --stanza=%s --delta restore", stanza)

	if err != nil {
		return fmt.Errorf("Error: %v\nstderr:\n%v\nstdout:\n%v\n", err, stdout, stderr)
	}

	err = ctl.cluster.Start()

	if err != nil {
		return err
	}

	return nil
}

// RestoreToPIT a specific point in time
func (ctl Controller) RestoreToPIT(stanza string, pointInTime time.Time) error {
	err := ctl.cluster.Stop()

	if err != nil {
		return err
	}

	stdout, stderr, err := ctl.runner.Run(
		"sudo -u postgres pgbackrest"+
			" --stanza=%s"+
			" --delta"+
			" --type=time"+
			" --target=\"%s\""+
			" restore",
		stanza,
		fmt.Sprintf((pointInTime.Format(time.RFC3339Nano))),
	)

	if err != nil {
		return fmt.Errorf("Error: %v\nstderr:\n%v\nstdout:\n%v\n", err, stdout, stderr)
	}

	err = ctl.cluster.Start()

	if err != nil {
		return err
	}

	return nil
}

// RestoreToSavePoint restores to the given savepoint
func (ctl Controller) RestoreToSavePoint(stanza string, savePoint string) error {
	err := ctl.cluster.Stop()

	if err != nil {
		return err
	}

	stdout, stderr, err := ctl.runner.Run(
		"sudo -u postgres pgbackrest"+
			" --stanza=%s"+
			" --delta"+
			" --type=name"+
			" --target=\"%s\""+
			" restore",
		stanza,
		savePoint,
	)

	if err != nil {
		return fmt.Errorf("Error: %v\nstderr:\n%v\nstdout:\n%v\n", err, stdout, stderr)
	}

	err = ctl.cluster.Start()

	if err != nil {
		return err
	}

	return nil
}

// RestoreToTransactionID restores to the given savepoint
func (ctl Controller) RestoreToTransactionID(stanza string, txId int64) error {
	err := ctl.cluster.Stop()

	if err != nil {
		return err
	}

	stdout, stderr, err := ctl.runner.Run(
		"sudo -u postgres pgbackrest"+
			" --stanza=%s"+
			" --delta"+
			" --type=xid"+
			" --target=\"%d\""+
			" restore",
		stanza,
		txId,
	)

	if err != nil {
		return fmt.Errorf("Error: %v\nstderr:\n%v\nstdout:\n%v\n", err, stdout, stderr)
	}

	err = ctl.cluster.Start()

	if err != nil {
		return err
	}

	return nil
}

func parseInfo(stdout string) ([]Info, error) {
	infos := make([]Info, 0)
	err := json.Unmarshal([]byte(stdout), &infos)
	return infos, err
}
