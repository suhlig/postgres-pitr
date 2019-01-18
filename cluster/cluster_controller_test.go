package cluster_test

import (
	"database/sql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/suhlig/postgres-pitr/cluster"
	clstr "github.com/suhlig/postgres-pitr/cluster"
	"github.com/suhlig/postgres-pitr/sshrunner"
)

var _ = Describe("Cluster Controller", func() {
	var ssh *sshrunner.Runner
	var cluster cluster.Controller

	BeforeEach(func() {
		ssh, err = ssh.New(masterHost)
		Expect(err).NotTo(HaveOccurred())

		cluster = clstr.NewController(ssh, config.Master.Version, config.Master.ClusterName)
	})

	It("knows the data directory", func() {
			Expect(cluster.DataDirectory()).To(Equal("/var/lib/postgresql/11/main"))
	})

	Context("a running cluster", func() {
		BeforeEach(func() {
			cluster.Start()
		})

		It("provides the status of the cluster", func() {
			running, err := cluster.IsRunning()
			Expect(err).NotTo(HaveOccurred())
			Expect(running).To(BeTrue())
		})

		It("can start the cluster", func() {
			err := cluster.Start()
			Expect(err).ToNot(HaveOccurred())
		})

		It("can stop the cluster", func() {
			err := cluster.Stop()
			Expect(err).ToNot(HaveOccurred())
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
	})

	Context("a stopped cluster", func() {
		BeforeEach(func() {
			cluster.Stop()
		})

		AfterEach(func() {
			cluster.Start()
		})

		It("provides the status of the cluster", func() {
			running, err := cluster.IsRunning()
			Expect(err).NotTo(HaveOccurred())
			Expect(running).To(BeFalse())
		})

		It("can start the cluster", func() {
			err := cluster.Start()
			Expect(err).ToNot(HaveOccurred())
		})

		It("can stop the cluster", func() {
			err := cluster.Stop()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("a non-existing cluster version", func() {
		BeforeEach(func() {
			cluster = clstr.NewController(ssh, "42", config.Master.ClusterName)
		})

		It("provides an error instead of the status of the cluster", func() {
			_, err := cluster.IsRunning()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("a non-existing cluster name", func() {
		BeforeEach(func() {
			cluster = clstr.NewController(ssh, config.Master.Version, "does-not-exist")
		})

		It("provides an error instead of the status of the cluster", func() {
			_, err := cluster.IsRunning()
			Expect(err).To(HaveOccurred())
		})
	})
})
