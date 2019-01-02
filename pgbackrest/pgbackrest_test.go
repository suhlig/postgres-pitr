package pgbackrest_test

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/suhlig/postgres-pitr/cluster"
	"github.com/suhlig/postgres-pitr/config"
	"github.com/suhlig/postgres-pitr/pgbackrest"
	"github.com/suhlig/postgres-pitr/sshrunner"
)

var _ = Describe("PgBackRest", func() {
	var config config.Config
	var err error

	BeforeEach(func() {
		config, err = config.FromFile("../config.yml")
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Database URL exists", func() {
		var masterURL string
		var masterDB *sql.DB

		BeforeEach(func() {
			masterURL, err = config.MasterDatabaseURL()
			Expect(err).NotTo(HaveOccurred())

			masterDB, err = sql.Open("postgres", masterURL)
			Expect(err).NotTo(HaveOccurred())
		})

		It("can be pinged", func() {
			err = masterDB.Ping()
			Expect(err).NotTo(HaveOccurred())
		})

		It("has the expected server version", func() {
			var version int
			err = masterDB.QueryRow("SHOW server_version_num;").Scan(&version)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(BeNumerically(">=", 110000))
		})
	})

	Context("with at least one base backup", func() {
		var masterSSH *sshrunner.Runner
		var masterPgBackRest pgbackrest.Controller
		var masterCluster cluster.Controller

		BeforeEach(func() {
			masterSSH, err = masterSSH.New(masterHost)
			Expect(err).NotTo(HaveOccurred())

			masterCluster = cluster.NewController(masterSSH, config.Master.Version, config.Master.ClusterName)
			masterPgBackRest = pgbackrest.NewController(masterSSH, masterCluster)

			err = masterPgBackRest.Backup(config.PgBackRest.Stanza)
			Expect(err).NotTo(HaveOccurred())
		})

		It("has info about the most recent backup", func() {
			infos, err := masterPgBackRest.Info(config.PgBackRest.Stanza)
			Expect(err).NotTo(HaveOccurred())

			Expect(infos).To(HaveLen(1))
			info := infos[0]
			Expect(info.Name).To(Equal(config.PgBackRest.Stanza))
			Expect(info.Status.Code).To(Equal(0), "Message is %s", info.Status.Message)
			Expect(info.Status.Message).To(Equal("ok"))
		})

		// https://pgbackrest.org/user-guide.html#quickstart/perform-restore
		When("an important file is lost", func() {
			It("restores the cluster", func() {
				By("deleting the pg_control file", func() {
					err = masterCluster.Stop()
					Expect(err).NotTo(HaveOccurred())

					stdout, stderr, err := masterSSH.Run("sudo -u postgres rm --force /var/lib/postgresql/%s/%s/global/pg_control", config.Master.Version, config.Master.ClusterName)
					Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
				})

				By("attempting to start the cluster again", func() {
					err = masterCluster.Start()
					Expect(err).To(HaveOccurred())
				})

				By("restoring the backup", func() {
					err = masterPgBackRest.Restore(config.PgBackRest.Stanza)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		// https://pgbackrest.org/user-guide.html#pitr
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

			When("remembering a point in time where everything was good", func() {
				var backupPointInTime time.Time

				BeforeEach(func() {
					err = masterDB.QueryRow("select current_timestamp").Scan(&backupPointInTime)
					Expect(err).NotTo(HaveOccurred())
				})

				It("can be restored to the given point in time", func() {
					By("dropping the important table", func() {
						_, err := masterDB.Exec("drop table important_table")
						Expect(err).NotTo(HaveOccurred())
					})

					By(fmt.Sprintf("restoring the cluster to the point in time when the data was good: %v", backupPointInTime), func() {
						err = masterPgBackRest.RestoreToPIT(config.PgBackRest.Stanza, backupPointInTime)
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

			When("creating a savepoint where everything was good", func() {
				var savePoint string

				BeforeEach(func() {
					savePoint = randomName()

					stdout, stderr, err := masterSSH.Run("sudo -u postgres psql -c \"select pg_create_restore_point('%s')\"", savePoint)
					Expect(err).ToNot(HaveOccurred(), "stderr: %v\nstdout:%v\n", stderr, stdout)
				})

				It("can be restored to the given savepoint", func() {
					By("dropping the important table", func() {
						_, err := masterDB.Exec("drop table important_table")
						Expect(err).NotTo(HaveOccurred())
					})

					By(fmt.Sprintf("restoring the cluster to the savepoint when the data was good: %v", savePoint), func() {
						err = masterPgBackRest.RestoreToSavePoint(config.PgBackRest.Stanza, savePoint)
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

				It("can be restored to the given transaction id", func() {
					By("dropping the important table", func() {
						_, err := masterDB.Exec("drop table important_table")
						Expect(err).NotTo(HaveOccurred())
					})

					By(fmt.Sprintf("restoring the cluster to the transaction id when the data was good: %v", txId), func() {
						err = masterPgBackRest.RestoreToTransactionID(config.PgBackRest.Stanza, txId)
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

			// https://pgbackrest.org/user-guide.html#replication/hot-standby
			Context("a hot standby exists", func() {
				var standbySSH *sshrunner.Runner
				var standbyPgBackRest pgbackrest.Controller
				var standbyCluster cluster.Controller
				var standbyURL string
				var standbyDB *sql.DB

				BeforeEach(func() {
					standbySSH, err = standbySSH.New(standbyHost)
					Expect(err).NotTo(HaveOccurred())

					standbyCluster = cluster.NewController(standbySSH, config.Standby.Version, config.Standby.ClusterName)
					standbyPgBackRest = pgbackrest.NewController(standbySSH, standbyCluster)

					standbyURL, err = config.StandbyDatabaseURL()
					Expect(err).NotTo(HaveOccurred())

					standbyDB, err = sql.Open("postgres", standbyURL)
					Expect(err).NotTo(HaveOccurred())
				})

				It("can be restored to provide the same data as the master", func() {
					By("creating a new backup of the master", func() {
						err = masterPgBackRest.Backup(config.PgBackRest.Stanza)
						Expect(err).NotTo(HaveOccurred())
					})

					By("restoring the backup on the standby", func() {
						err = standbyPgBackRest.Restore(config.PgBackRest.Stanza)
						Expect(err).NotTo(HaveOccurred())
					})

					By(fmt.Sprintf("checking that the important data '%s' exists on the hot standby", importantData), func() {
						var count int
						err = standbyDB.QueryRow("select count(message) from important_table where message = $1", importantData).Scan(&count)
						Expect(err).NotTo(HaveOccurred())
						Expect(count).To(Equal(1))
					})
				})

				XIt("does not allow writes (read-only)", func() {

				})

				XIt("allows subsequent restores while running", func() {

				})
			})
		})
	})
})
