package cluster_test

import (
	"database/sql"

	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/suhlig/postgres-pitr/cluster"
	clstr "github.com/suhlig/postgres-pitr/cluster"
	"github.com/suhlig/postgres-pitr/sshrunner"
)

var _ = Describe("Cluster", func() {
	var ssh *sshrunner.Runner
	var cluster cluster.Controller

	BeforeEach(func() {
		ssh, err = ssh.New(masterHost)
		Expect(err).NotTo(HaveOccurred())

		cluster = clstr.NewController(ssh, config.Master.Version, config.Master.ClusterName)
	})

	Context("is running", func() {
		var masterURL string
		var masterDB *sql.DB

		BeforeEach(func() {
			cluster.Start()

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
