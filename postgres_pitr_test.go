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

	Context("with at least one base backup", func() {
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
			err = pgbr.Backup(config.PgBackRest.Stanza)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			forceRestore(pgbr, clustr, config.PgBackRest.Stanza)
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

		// https://pgbackrest.org/user-guide.html#pitr
		Context("important data exists", func() {
			var url string
			var db *sql.DB

			BeforeEach(func() {
				url, err = config.DatabaseURL()
				Expect(err).NotTo(HaveOccurred())

				db, err = sql.Open("postgres", url)
				Expect(err).NotTo(HaveOccurred())

				_, err := db.Exec("create table IF NOT EXISTS important_table (message text)")
				Expect(err).NotTo(HaveOccurred())

				_, err = db.Exec("insert into important_table values ($1)", "Important Data")
				Expect(err).NotTo(HaveOccurred())
			})

			When("remembering a point in time where everything was good", func() {
				var backupPointInTime time.Time

				BeforeEach(func() {
					err = db.QueryRow("select current_timestamp").Scan(&backupPointInTime)
					Expect(err).NotTo(HaveOccurred())
				})

				It("can be restored to the given point in time", func() {
					By("dropping the important table", func() {
						_, err := db.Exec("drop table important_table")
						Expect(err).NotTo(HaveOccurred())
					})

					By(fmt.Sprintf("restoring the cluster to the point in time when the data was good: %v", backupPointInTime), func() {
						err = clustr.Stop()
						Expect(err).NotTo(HaveOccurred())

						err := clustr.Clear()
						Expect(err).NotTo(HaveOccurred())

						err = pgbr.RestoreToPIT(config.PgBackRest.Stanza, backupPointInTime)
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

					By("Check that the important table exists", func() {
						var message string
						err = db.QueryRow("select message from important_table").Scan(&message)
						Expect(err).NotTo(HaveOccurred())
						Expect(message).To(Equal("Important Data"))
					})
				})
			})

			When("creating a savepoint where everything was good", func() {
				var savePoint string

				BeforeEach(func() {
					savePoint = randomName()

					stdout, stderr, err := ssh.Run("sudo -u postgres psql -c \"select pg_create_restore_point('%s')\"", savePoint)
					Expect(err).ToNot(HaveOccurred(), "stderr: %v\nstdout:%v\n", stderr, stdout)
				})

				It("can be restored to the given savepoint", func() {
					By("dropping the important table", func() {
						_, err := db.Exec("drop table important_table")
						Expect(err).NotTo(HaveOccurred())
					})

					By(fmt.Sprintf("restoring the cluster to the savepoint when the data was good: %v", savePoint), func() {
						err = clustr.Stop()
						Expect(err).NotTo(HaveOccurred())

						err := clustr.Clear()
						Expect(err).NotTo(HaveOccurred())

						err = pgbr.RestoreToSavePoint(config.PgBackRest.Stanza, savePoint)
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

					By("Check that the important table exists", func() {
						var message string
						err = db.QueryRow("select message from important_table").Scan(&message)
						Expect(err).NotTo(HaveOccurred())
						Expect(message).To(Equal("Important Data"))
					})
				})
			})

			When("saving a transaction id where everything was good", func() {
				XIt("can be restored to the given transaction id", func() {
				})
			})
		})
	})
})
