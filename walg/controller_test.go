package walg_test

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"

	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/suhlig/postgres-pitr/cluster"
	"github.com/suhlig/postgres-pitr/config"
	"github.com/suhlig/postgres-pitr/sshrunner"
	"github.com/suhlig/postgres-pitr/walg"
)

var _ = Describe("WAL-G controller", func() {
	var config config.Config
	var err error

	BeforeEach(func() {
		config, err = config.FromFile("../config.yml")
		Expect(err).NotTo(HaveOccurred())
	})

	Context("with at least one base backup", func() {
		var ssh *sshrunner.Runner
		var wlg walg.Controller
		var masterCluster cluster.Controller

		BeforeEach(func() {
			ssh, err = ssh.New(masterHost)
			Expect(err).NotTo(HaveOccurred())

			masterCluster = cluster.NewController(ssh, config.Master.Version, config.Master.ClusterName)
			wlg = walg.NewController(ssh, masterCluster)
		})

		JustBeforeEach(func() {
			err = wlg.Backup()
			Expect(err).NotTo(HaveOccurred())
		})

		It("has info at least one backup", func() {
			backups, err := wlg.List()
			Expect(err).NotTo(HaveOccurred())
			Expect(backups.Path).ToNot(BeEmpty())
			Expect(backups.Backups).ToNot(BeEmpty())
		})

		When("an important file is lost", func() {
			It("restores the cluster", func() {
				By("deleting the pg_control file", func() {
					err = masterCluster.Stop()
					Expect(err).NotTo(HaveOccurred())

					stdout, stderr, err := ssh.Run("sudo --user postgres rm --force /var/lib/postgresql/%s/%s/global/pg_control", config.Master.Version, config.Master.ClusterName)
					Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
				})

				By("attempting to start the cluster again", func() {
					err = masterCluster.Start()
					Expect(err).To(HaveOccurred())
				})

				By("restoring the backup", func() {
					// WIP Fails to start, probably because backup_label is still there and recovery.conf does not exist
					err = wlg.RestoreLatest()
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("important data exists", func() {
			var masterURL string
			var masterDB *sql.DB
			var importantData string

			BeforeEach(func() {
				masterURL, err = config.MasterDatabaseURL()
				Expect(err).NotTo(HaveOccurred())

				masterDB, err = sql.Open("postgres", masterURL)
				Expect(err).NotTo(HaveOccurred())

				_, err := masterDB.Exec("create table IF NOT EXISTS important_table (message text)")
				Expect(err).NotTo(HaveOccurred())

				importantData = randomName()
				_, err = masterDB.Exec("insert into important_table values ($1)", importantData)
				Expect(err).NotTo(HaveOccurred())
			})

			When("saving a transaction id where everything was good", func() {
				var txId int64

				BeforeEach(func() {
					tx, err := masterDB.Begin()
					Expect(err).NotTo(HaveOccurred())

					err = tx.QueryRow("select txid_current()").Scan(&txId)
					Expect(err).NotTo(HaveOccurred())
					Expect(txId).NotTo(BeNil())

					importantData = randomName() // make sure we use a new value while in Tx
					_, err = tx.Exec("insert into important_table values ($1)", importantData)
					Expect(err).NotTo(HaveOccurred())

					err = tx.Commit()
					Expect(err).NotTo(HaveOccurred())
				})

				FIt("can be restored to the given transaction id", func() {
					By("dropping the important table", func() {
						_, err := masterDB.Exec("drop table important_table")
						Expect(err).NotTo(HaveOccurred())
					})

					By(fmt.Sprintf("restoring the cluster to the transaction id when the data was good: %v", txId), func() {
						err = wlg.RestoreToTransactionID(txId)
						Expect(err).NotTo(HaveOccurred())
					})

					By(fmt.Sprintf("checking that the important data '%s' exists", importantData), func() {
						var count int
						err = masterDB.QueryRow("select count(message) from important_table where message = $1", importantData).Scan(&count)
						Expect(err).NotTo(HaveOccurred())
						Expect(count).To(Equal(1))
					})
				})
			})
		})
	})
})

// TODO This is a copy
func randomName() string {
	chars := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	length := 16
	var builder strings.Builder

	for i := 0; i < length; i++ {
		builder.WriteRune(chars[rand.Intn(len(chars))])
	}

	return builder.String()
}
