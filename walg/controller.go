package walg

import (
	"time"

	pitr "github.com/suhlig/postgres-pitr"
	"github.com/suhlig/postgres-pitr/cluster"
)

// Controller provides a way to control WAL-G
type Controller struct {
	runner  pitr.Runner
	cluster cluster.Controller
}

// Backup describes a single PostgreSQL backup made by WAL-G
type Backup struct {
	Name                  string
	LastModified          time.Time
	WalSegmentBackupStart string
}

// Info provides all details about what is backed up by WAL-G
type Info struct {
	Path    string
	Backups []Backup
}

// NewController creates a new controller with the given runner and cluster controller
func NewController(runner pitr.Runner, cluster cluster.Controller) Controller {
	return Controller{
		runner:  runner,
		cluster: cluster,
	}
}

// Backup creates a new backup for the given cluster version and -name
func (ctl Controller) Backup() *pitr.Error {
	stdout, stderr, err := ctl.runner.Run("sudo -u postgres /opt/wal-g/bin/base-backup /var/lib/postgresql/%s/%s", ctl.cluster.Version, ctl.cluster.Name)

	if err != nil {
		return &pitr.Error{
			Message: err.Error(),
			Stdout:  stdout,
			Stderr:  stderr,
		}
	}

	return nil
}

// Restore the latest backup
func (ctl Controller) RestoreLatest() *pitr.Error {
	return ctl.Restore("LATEST")
}

// Restore the backup with the given name
func (ctl Controller) Restore(name string) *pitr.Error {
	err := ctl.cluster.Stop()

	if err != nil {
		return err
	}

	err = ctl.cluster.Clear()

	if err != nil {
		return err
	}

	stdout, stderr, runErr := ctl.runner.Run("sudo --login --user postgres wal-g backup-fetch /var/lib/postgresql/%s/%s %s", ctl.cluster.Version, ctl.cluster.Name, name)

	if runErr != nil {
		return &pitr.Error{
			Message: runErr.Error(),
			Stdout:  stdout,
			Stderr:  stderr,
		}
	}

	err = ctl.cluster.Start()

	if err != nil {
		return err
	}

	return nil
}

// RestoreToTransactionID restores to the given savepoint
func (ctl Controller) RestoreToTransactionID(txID int64) *pitr.Error {
	err := ctl.cluster.Stop()

	if err != nil {
		return err
	}

	pgDataDir := fmt.Sprintf("/var/lib/postgresql/%s/%s", ctl.cluster.Version, ctl.cluster.Name)
	echoCmd := `echo "restore_command = 'bash --login -c \"wal-g wal-fetch %f %p\"'"`
	stdout, stderr, runErr := ctl.runner.Run("%s | sudo --login --user postgres tee %s/recovery.conf", echoCmd, pgDataDir)

	echoCmd = fmt.Sprintf("echo recovery_target_xid = %d", txID)
	stdout, stderr, runErr = ctl.runner.Run("%s | sudo --login --user postgres tee --append %s/recovery.conf", echoCmd, pgDataDir)

	if runErr != nil {
		return &pitr.Error{
			Message: runErr.Error(),
			Stdout:  stdout,
			Stderr:  stderr,
		}
	}

	err = ctl.cluster.Start()

	if err != nil {
		return err
	}

	return nil
}

// List provides a summary of backups
func (ctl Controller) List() (*Info, *pitr.Error) {
	stdout, stderr, err := ctl.runner.Run("sudo -u postgres /opt/wal-g/bin/list-backups")

	if err != nil {
		return nil, &pitr.Error{
			Message: err.Error(),
			Stdout:  stdout,
			Stderr:  stderr,
		}
	}

	infos, err := ParseListOutput(stdout)

	if err != nil {
		return nil, &pitr.Error{
			Message: "Parse error",
			Stdout:  stderr,
			Stderr:  stderr,
		}
	}

	return infos, nil
}
