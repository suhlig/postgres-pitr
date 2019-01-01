package postgres_pitr_test

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"

	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/suhlig/postgres-pitr/cluster"
	"github.com/suhlig/postgres-pitr/config"
	"github.com/suhlig/postgres-pitr/pgbackrest"
	"github.com/suhlig/postgres-pitr/sshrunner"
)

// Quickly restore from a backup
// This may also be useful to run as one-off in a Before function
func forceRestore(pgbr pgbackrest.Controller, clustr cluster.Controller, stanza string) {
	clustr.Stop()

	err := clustr.Clear()
	Expect(err).NotTo(HaveOccurred())

	err = pgbr.Restore(stanza)
	Expect(err).NotTo(HaveOccurred())

	err = clustr.Start()
	Expect(err).NotTo(HaveOccurred())
}

func randomName() string {
	chars := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	length := 16
	var builder strings.Builder

	for i := 0; i < length; i++ {
		builder.WriteRune(chars[rand.Intn(len(chars))])
	}

	return builder.String()
}

var _ = Describe("a PostgreSQL cluster", func() {
	var config config.Config
	var err error

	BeforeEach(func() {
		config, err = config.FromFile("config.yml")
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

			masterCluster, err = cluster.NewController(masterSSH, config.Master.Version, config.Master.ClusterName)
			Expect(err).NotTo(HaveOccurred())

			masterPgBackRest, err = pgbackrest.NewController(masterSSH)
			Expect(err).NotTo(HaveOccurred())
			err = masterPgBackRest.Backup(config.PgBackRest.Stanza)
			Expect(err).NotTo(HaveOccurred())
		})

		// AfterEach(func() {
		// 	forceRestore(masterPgBackRest, masterCluster, config.PgBackRest.Stanza)
		// })

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
				By("stopping the cluster", func() {
					err = masterCluster.Stop()
					Expect(err).NotTo(HaveOccurred())
				})

				By("deleting the pg_control file", func() {
					stdout, stderr, err := masterSSH.Run("sudo -u postgres rm --force /var/lib/postgresql/%s/%s/global/pg_control", config.Master.Version, config.Master.ClusterName)
					Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
				})

				By("attempting to start the cluster again", func() {
					err = masterCluster.Start()
					Expect(err).To(HaveOccurred())
				})

				By("removing all files from the PostgreSQL data directory", func() {
					err = masterCluster.Clear()
					Expect(err).ToNot(HaveOccurred())
				})

				By("restoring the backup", func() {
					err = masterPgBackRest.Restore(config.PgBackRest.Stanza)
					Expect(err).NotTo(HaveOccurred())
				})

				By("starting the cluster", func() {
					err = masterCluster.Start()
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		// https://pgbackrest.org/user-guide.html#pitr
		Context("important data exists", func() {
			var masterURL string
			var masterDB *sql.DB

			BeforeEach(func() {
				masterURL, err = config.MasterDatabaseURL()
				Expect(err).NotTo(HaveOccurred())

				masterDB, err = sql.Open("postgres", masterURL)
				Expect(err).NotTo(HaveOccurred())

				_, err := masterDB.Exec("create table IF NOT EXISTS important_table (message text)")
				Expect(err).NotTo(HaveOccurred())

				_, err = masterDB.Exec("insert into important_table values ($1)", "Important Data")
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
						err = masterCluster.Stop()
						Expect(err).NotTo(HaveOccurred())

						err := masterCluster.Clear()
						Expect(err).NotTo(HaveOccurred())

						err = masterPgBackRest.RestoreToPIT(config.PgBackRest.Stanza, backupPointInTime)
						Expect(err).NotTo(HaveOccurred())
					})

					By("displaying recovery.conf", func() {
						stdout, stderr, err := masterSSH.Run("sudo -u postgres cat /var/lib/postgresql/%s/%s/recovery.conf", config.Master.Version, config.Master.ClusterName)
						println(stdout)
						Expect(err).ToNot(HaveOccurred(), "stderr: %v\nstdout:%v\n", stderr, stdout)
					})

					By("starting the cluster", func() {
						err = masterCluster.Start()
						Expect(err).NotTo(HaveOccurred())
					})

					By("checking that the important data exists", func() {
						var message string
						err = masterDB.QueryRow("select message from important_table").Scan(&message)
						Expect(err).NotTo(HaveOccurred())
						Expect(message).To(Equal("Important Data"))
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
						err = masterCluster.Stop()
						Expect(err).NotTo(HaveOccurred())

						err := masterCluster.Clear()
						Expect(err).NotTo(HaveOccurred())

						err = masterPgBackRest.RestoreToSavePoint(config.PgBackRest.Stanza, savePoint)
						Expect(err).NotTo(HaveOccurred())
					})

					By("displaying recovery.conf", func() {
						stdout, stderr, err := masterSSH.Run("sudo -u postgres cat /var/lib/postgresql/%s/%s/recovery.conf", config.Master.Version, config.Master.ClusterName)
						println(stdout)
						Expect(err).ToNot(HaveOccurred(), "stderr: %v\nstdout:%v\n", stderr, stdout)
					})

					By("starting the cluster", func() {
						err = masterCluster.Start()
						Expect(err).NotTo(HaveOccurred())
					})

					By("checking that the important data exists", func() {
						var message string
						err = masterDB.QueryRow("select message from important_table").Scan(&message)
						Expect(err).NotTo(HaveOccurred())
						Expect(message).To(Equal("Important Data"))
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

					_, err = tx.Exec("insert into important_table values ($1)", "Important Data")
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
						err = masterCluster.Stop()
						Expect(err).NotTo(HaveOccurred())

						err := masterCluster.Clear()
						Expect(err).NotTo(HaveOccurred())

						err = masterPgBackRest.RestoreToTransactionId(config.PgBackRest.Stanza, txId)
						Expect(err).NotTo(HaveOccurred())
					})

					By("displaying recovery.conf", func() {
						stdout, stderr, err := masterSSH.Run("sudo -u postgres cat /var/lib/postgresql/%s/%s/recovery.conf", config.Master.Version, config.Master.ClusterName)
						println(stdout)
						Expect(err).ToNot(HaveOccurred(), "stderr: %v\nstdout:%v\n", stderr, stdout)
					})

					By("starting the cluster", func() {
						err = masterCluster.Start()
						Expect(err).NotTo(HaveOccurred())
					})

					By("checking that the important data exists", func() {
						var message string
						err = masterDB.QueryRow("select message from important_table").Scan(&message)
						Expect(err).NotTo(HaveOccurred())
						Expect(message).To(Equal("Important Data"))
					})
				})
			})

			// https://pgbackrest.org/user-guide.html#replication/hot-standby
			Context("a hot standby is restored from scratch", func() {
				var standbySSH *sshrunner.Runner
				var standbyPgBackRest pgbackrest.Controller
				var standbyCluster cluster.Controller
				var standbyURL string
				var standbyDB *sql.DB

				BeforeEach(func() {
					standbySSH, err = standbySSH.New(standbyHost)
					Expect(err).NotTo(HaveOccurred())

					standbyCluster, err = cluster.NewController(standbySSH, config.Standby.Version, config.Standby.ClusterName)
					Expect(err).NotTo(HaveOccurred())

					standbyPgBackRest, err = pgbackrest.NewController(standbySSH)
					Expect(err).NotTo(HaveOccurred())

					standbyURL, err = config.StandbyDatabaseURL()
					Expect(err).NotTo(HaveOccurred())

					standbyDB, err = sql.Open("postgres", standbyURL)
					Expect(err).NotTo(HaveOccurred())
				})

				It("provides the important information", func() {
					By("stopping the cluster", func() {
						err = standbyCluster.Stop()
						Expect(err).NotTo(HaveOccurred())
					})

					By("removing all files from the PostgreSQL data directory", func() {
						err = standbyCluster.Clear()
						Expect(err).ToNot(HaveOccurred())
					})

					By("restoring the backup", func() {
						err = standbyPgBackRest.Restore(config.PgBackRest.Stanza)
						Expect(err).NotTo(HaveOccurred())
					})

					By("starting the cluster", func() {
						err = standbyCluster.Start()
						Expect(err).NotTo(HaveOccurred())
					})

					By("checking that the important data exists", func() {
						Eventually(func() string {
							var message string
							err = standbyDB.QueryRow("select message from important_table").Scan(&message)
							// Expect(err).NotTo(HaveOccurred())
							return message
						}, "20s", "1s").Should(Equal("Important Data"))
					})
				})
			})
		})
	})
})
