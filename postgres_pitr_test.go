package postgres_pitr_test

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

var _ = Describe("a VM with PostgreSQL", func() {
	var config config.Config
	var err error

	BeforeEach(func() {
		config, err = config.New("config.yml")
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Database URL exists", func() {
		var url string
		var db *sql.DB

		BeforeEach(func() {
			url, err = config.DatabaseURL()
			Expect(err).NotTo(HaveOccurred())

			db, err = sql.Open("postgres", url)
			Expect(err).NotTo(HaveOccurred())
		})

		It("can be pinged", func() {
			err = db.Ping()
			Expect(err).NotTo(HaveOccurred())
		})

		It("has the expected server version", func() {
			var version int
			err = db.QueryRow("SHOW server_version_num;").Scan(&version)
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(BeNumerically(">=", 110000))
		})
	})

	Context("with SSH config", func() {
		var ssh *sshrunner.Runner
		var pgbr pgbackrest.Controller
		var clustr cluster.Controller

		BeforeEach(func() {
			ssh, err = ssh.New(host)
			Expect(err).NotTo(HaveOccurred())

			clustr, err = cluster.NewController(ssh, config.DB.Version, config.DB.ClusterName)
			Expect(err).NotTo(HaveOccurred())

			pgbr, err = pgbackrest.NewController(ssh)
			Expect(err).NotTo(HaveOccurred())
		})

		// Use this to quickly restore from a backup
		// FIt("DEBUG Restores a backup", func() {
		// 	clustr.Stop()

		// 	err = clustr.Clear()
		// 	Expect(err).NotTo(HaveOccurred())

		// 	err = pgbr.Restore(config.PgBackRest.Stanza)
		// 	Expect(err).NotTo(HaveOccurred())

		// 	err = clustr.Start()
		// 	Expect(err).NotTo(HaveOccurred())
		// })

		Context("A backup exists", func() {
			BeforeEach(func() {
				err = pgbr.Backup(config.PgBackRest.Stanza)
				Expect(err).NotTo(HaveOccurred())
			})

			It("has info about the most recent backup", func() {
				infos, err := pgbr.Info(config.PgBackRest.Stanza)
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
						err = clustr.Stop()
						Expect(err).NotTo(HaveOccurred())
					})

					By("deleting the pg_control file", func() {
						stdout, stderr, err := ssh.Run("sudo -u postgres rm --force /var/lib/postgresql/%s/%s/global/pg_control", config.DB.Version, config.DB.ClusterName)
						Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
					})

					By("attempting to start the cluster again", func() {
						err = clustr.Start()
						Expect(err).To(HaveOccurred())
					})

					By("removing all files from the PostgreSQL data directory", func() {
						err = clustr.Clear()
						Expect(err).ToNot(HaveOccurred())
					})

					By("restoring the backup", func() {
						err = pgbr.Restore(config.PgBackRest.Stanza)
						Expect(err).NotTo(HaveOccurred())
					})

					By("starting the cluster", func() {
						err = clustr.Start()
						Expect(err).NotTo(HaveOccurred())
					})
				})
			})
		})

		// https://pgbackrest.org/user-guide.html#pitr
		Context("important data was deleted", func() {
			var url string
			var db *sql.DB
			var pgbr pgbackrest.Controller

			BeforeEach(func() {
				url, err = config.DatabaseURL()
				Expect(err).NotTo(HaveOccurred())

				db, err = sql.Open("postgres", url)
				Expect(err).NotTo(HaveOccurred())

				pgbr, err = pgbackrest.NewController(ssh)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				_, err := db.Exec("drop table important_table")
				Expect(err).NotTo(HaveOccurred())
			})

			It("can be restored to the state at the given point in time", func() {
				var backupPointInTime time.Time

				By("creating a table with very important data", func() {
					_, err := db.Exec("create table IF NOT EXISTS important_table (message text)")
					Expect(err).NotTo(HaveOccurred())

					_, err = db.Exec("insert into important_table values ($1)", "Important Data")
					Expect(err).NotTo(HaveOccurred())
				})

				By("backing up the demo cluster", func() {
					err = pgbr.Backup(config.PgBackRest.Stanza)
					Expect(err).NotTo(HaveOccurred())
				})

				By("remembering a point in time where everything was good", func() {
					err = db.QueryRow("select current_timestamp").Scan(&backupPointInTime)
					Expect(err).NotTo(HaveOccurred())
				})

				By("dropping the important table", func() {
					_, err := db.Exec("drop table important_table")
					Expect(err).NotTo(HaveOccurred())
				})

				By(fmt.Sprintf("restoring the cluster to the point in time when it was backed up: %v", backupPointInTime), func() {
					err = clustr.Stop()
					Expect(err).NotTo(HaveOccurred())

					err = pgbr.RestoreTo(config.PgBackRest.Stanza, backupPointInTime)
					Expect(err).NotTo(HaveOccurred())
				})

				By("displaying recovery.conf", func() {
					stdout, stderr, err := ssh.Run("sudo -u postgres cat /var/lib/postgresql/%s/%s/recovery.conf", config.DB.Version, config.DB.ClusterName)
					println(stdout)
					Expect(err).ToNot(HaveOccurred(), "stderr: %v\nstdout:%v\n", stderr, stdout)
				})

				By("starting the cluster", func() {
					err = clustr.Start()
					Expect(err).NotTo(HaveOccurred())
				})

				// TODO Without pg_wal_replay_resume(), the test blocks forever because the DB is read-only
				By("resuming WAL replay", func() {
					stdout, stderr, err := ssh.Run("sudo -u postgres psql -c 'select pg_wal_replay_resume();'")
					Expect(err).ToNot(HaveOccurred(), "stderr: %v\nstdout:%v\n", stderr, stdout)
				})

				By("Check that the important table exists", func() {
					var message string
					err = db.QueryRow("select message from important_table").Scan(&message)
					Expect(err).NotTo(HaveOccurred())
					Expect(message).To(Equal("Important Data"))
				})
			})
		})
	})
})
