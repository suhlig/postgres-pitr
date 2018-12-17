package postgres_pitr_test

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/mikkeloscar/sshconfig"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/suhlig/postgres-pitr/config"
	"github.com/suhlig/postgres-pitr/pgbackrest"
	"github.com/suhlig/postgres-pitr/vagrant"
)

var _ = Describe("VM with pgBackRest", func() {
	var config config.Config
	var vagrant *vagrant.VagrantSSH
	var err error
	var url string

	BeforeEach(func() {
		config, err = config.New("config.yml")
		Expect(err).NotTo(HaveOccurred())

		url, err = config.DatabaseURL()
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Database connection established", func() {
		var db *sql.DB

		BeforeEach(func() {
			db, err = sql.Open("postgres", url)
			Expect(err).NotTo(HaveOccurred())
		})

		It("can connect", func() {
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

	Context("Config file exists", func() {
		Context("for the cluster", func() {
			It("has the configured server version", func() {
				Expect(config.DB.Version).To(Equal("11"))
			})

			It("has the configured cluster name", func() {
				Expect(config.DB.ClusterName).To(Equal("main"))
			})
		})

		Context("for pgbackrest", func() {
			It("has the configured stanza", func() {
				Expect(config.PgBackRest.Stanza).To(Equal("pitr"))
			})
		})
	})

	Context("VM has SSH config", func() {
		BeforeEach(func() {
			hosts, err := sshconfig.ParseSSHConfig(configFileName)
			Expect(len(hosts)).To(BeNumerically("==", 1), "Require exactly one host, but found %d", len(hosts))

			vagrant, err = vagrant.New(*hosts[0])
			Expect(err).NotTo(HaveOccurred())
		})

		It("can connect using SSH", func() {
			stdout, stderr, err := vagrant.Run("id")
			Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
			Expect(stdout).To(ContainSubstring("vagrant"))
		})

		It("can run commands with args via SSH", func() {
			stdout, stderr, err := vagrant.Run("ls -l")
			Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
			Expect(stdout).To(ContainSubstring("total"))
		})

		Context("A backup exists", func() {
			BeforeEach(func() {
				stdout, stderr, err := vagrant.Run("sudo -u postgres pgbackrest --stanza=%s backup", config.PgBackRest.Stanza)
				Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
			})

			It("lists the backup", func() {
				stdout, stderr, err := vagrant.Run("sudo -u postgres pgbackrest info --stanza=%s --output=json", config.PgBackRest.Stanza)
				Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)

				infos, err := pgbackrest.ParseInfo(stdout)
				Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)

				Expect(infos).To(HaveLen(1))
				info := infos[0]
				Expect(info.Name).To(Equal(config.PgBackRest.Stanza))
				Expect(info.Status.Code).To(Equal(0), "Message is %s", info.Status.Message)
				Expect(info.Status.Message).To(Equal("ok"))
			})

			XWhen("the pg_control file is lost", func() {
				It("successfully restores the cluster", func() {

				})
			})

			XWhen("important data was deleted", func() {
				It("can be restored to the state at the given point in time", func() {
					// TODO https://pgbackrest.org/user-guide.html#pitr
				})
			})

			/*
				// Verify cluster status:
				sudo pg_ctlcluster 11 main status

				// Stop the main cluster and delete the pg_control file
				sudo pg_ctlcluster 9.4 main stop
				sudo -u postgres rm /var/lib/postgresql/9.4/main/global/pg_control

				// Verify that starting the cluster without this important file will result in an error
				sudo pg_ctlcluster 9.4 main start

				// ensure cluster is stopped
				sudo pg_ctlcluster 11 main status

				// remove all files from the PostgreSQL data directory
				sudo -u postgres find /var/lib/postgresql/9.4/main -mindepth 1 -delete

				// Restore the main cluster and start PostgreSQL
				sudo -u postgres pgbackrest --stanza=pitr restore
				sudo pg_ctlcluster 9.4 main start

				// Verify that the cluster is up
			*/
		})
	})
})
