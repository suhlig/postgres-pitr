package postgres_pitr_test

import (
	"database/sql"

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

		BeforeEach(func() {
			ssh, err = ssh.New(host)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("A backup exists", func() {
			var pgbr pgbackrest.Controller

			BeforeEach(func() {
				pgbr, err = pgbackrest.NewController(ssh)
				Expect(err).NotTo(HaveOccurred())

				err = pgbr.Backup(config.PgBackRest.Stanza)
				Expect(err).NotTo(HaveOccurred())
			})

			It("has a backup", func() {
				infos, err := pgbr.Info(config.PgBackRest.Stanza)
				Expect(err).NotTo(HaveOccurred())

				Expect(infos).To(HaveLen(1))
				info := infos[0]
				Expect(info.Name).To(Equal(config.PgBackRest.Stanza))
				Expect(info.Status.Code).To(Equal(0), "Message is %s", info.Status.Message)
				Expect(info.Status.Message).To(Equal("ok"))
			})

			When("an important file is lost", func() {
				var clustr cluster.Controller

				BeforeEach(func() {
					clustr, err = cluster.NewController(ssh)
					Expect(err).NotTo(HaveOccurred())
				})

				It("restores the cluster", func() {
					By("stopping the cluster", func() {
						err = clustr.Stop(config.DB.Version, config.DB.ClusterName)
						Expect(err).NotTo(HaveOccurred())
					})

					By("deleting the pg_control file", func() {
						stdout, stderr, err := ssh.Run("sudo -u postgres rm /var/lib/postgresql/%s/%s/global/pg_control", config.DB.Version, config.DB.ClusterName)
						Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
					})

					By("attempting to start the cluster again", func() {
						err = clustr.Start(config.DB.Version, config.DB.ClusterName)
						Expect(err).To(HaveOccurred())
					})

					By("removing all files from the PostgreSQL data directory", func() {
						stdout, stderr, err := ssh.Run("sudo -u postgres find /var/lib/postgresql/%s/%s -mindepth 1 -delete", config.DB.Version, config.DB.ClusterName)
						Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
					})

					By("restoring the backup", func() {
						err = pgbr.Restore(config.PgBackRest.Stanza)
						Expect(err).NotTo(HaveOccurred())
					})

					By("starting the cluster", func() {
						err = clustr.Start(config.DB.Version, config.DB.ClusterName)
						Expect(err).NotTo(HaveOccurred())
					})
				})
			})

			XWhen("important data was deleted", func() {
				It("can be restored to the state at the given point in time", func() {
					// TODO https://pgbackrest.org/user-guide.html#pitr
				})
			})
		})
	})
})
