package walg_test

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/suhlig/postgres-pitr/cluster"
	"github.com/suhlig/postgres-pitr/config"
	h "github.com/suhlig/postgres-pitr/helpers"
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
		var masterURL string
		var masterDB *sql.DB

		BeforeEach(func() {
			ssh, err = ssh.New(masterHost)
			Expect(err).NotTo(HaveOccurred())

			masterCluster = cluster.NewController(ssh, config.Master.Version, config.Master.ClusterName)
			wlg = walg.NewController(ssh, masterCluster)

			err = wlg.Backup()
			Expect(err).NotTo(HaveOccurred())

			masterURL, err = config.MasterDatabaseURL()
			Expect(err).NotTo(HaveOccurred())

			masterDB, err = sql.Open("postgres", masterURL)
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

					stdout, stderr, err := ssh.Run("sudo --user postgres rm --force %s/global/pg_control", masterCluster.DataDirectory())
					Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
				})

				By("attempting to start the cluster again", func() {
					err = masterCluster.Start()
					Expect(err).To(HaveOccurred())
				})

				By("restoring the backup", func() {
					err = wlg.RestoreLatest()
					Expect(err).NotTo(HaveOccurred())
				})

				By("checking that the restored database is writable", func() {
					var isReadOnly bool
					err = masterDB.QueryRow("select pg_is_in_recovery()").Scan(&isReadOnly)
					Expect(err).NotTo(HaveOccurred())
					Expect(isReadOnly).To(BeFalse())
				})
			})
		})

		Context("important data exists", func() {
			var importantData string

			BeforeEach(func() {
				_, err := masterDB.Exec("create table IF NOT EXISTS important_table (message text)")
				Expect(err).NotTo(HaveOccurred())
			})

			When("saving a transaction id where everything was good", func() {
				var txId int64
				var walID string

				BeforeEach(func() {
					tx, err := masterDB.Begin()
					Expect(err).NotTo(HaveOccurred())

					err = tx.QueryRow("select txid_current()").Scan(&txId)
					Expect(err).NotTo(HaveOccurred())
					Expect(txId).NotTo(BeNil())

					importantData = h.RandomName() // make sure we use a new value while in Tx
					_, err = tx.Exec("insert into important_table values ($1)", importantData)
					Expect(err).NotTo(HaveOccurred())

					err = tx.Commit()
					Expect(err).NotTo(HaveOccurred())

					err = masterDB.QueryRow("select pg_switch_wal()").Scan(&walID)
					Expect(err).NotTo(HaveOccurred())
				})

				It("can be restored to the given transaction id", func() {
					By("dropping the important table", func() {
						_, err := masterDB.Exec("drop table important_table")
						Expect(err).NotTo(HaveOccurred())
					})

					By(fmt.Sprintf("restoring the cluster to the transaction id when the data was good: %v", txId), func() {
						err = wlg.RestoreToTransactionID(txId)
						Expect(err).NotTo(HaveOccurred())
					})

					By(fmt.Sprintf("checking that the important data '%s' exists (archived as %s)", importantData, walID), func() {
						var count int
						err = masterDB.QueryRow("select count(message) from important_table where message = $1", importantData).Scan(&count)
						Expect(err).NotTo(HaveOccurred())
						Expect(count).To(Equal(1))
					})

					By("checking that we can write to the restored database", func() {
						_, err = masterDB.Exec("insert into important_table values ($1)", h.RandomName())
						Expect(err).NotTo(HaveOccurred())
					})
				})
			})
		})
	})
})
