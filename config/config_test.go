package config_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/suhlig/postgres-pitr/config"
)

var _ = Describe("Config", func() {
	var config config.Config
	var err error

	BeforeEach(func() {
		config, err = config.FromFile("../config.yml")
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Configuration file contains an entry", func() {
		Context("for the master cluster", func() {
			It("has the configured server version", func() {
				Expect(config.Master.Version).ToNot(BeEmpty())
				Expect(config.Master.Version).To(Equal("11"))
			})

			It("has the configured master cluster name", func() {
				Expect(config.Master.ClusterName).ToNot(BeEmpty())
				Expect(config.Master.ClusterName).To(Equal("main"))
			})

			It("has the database URL", func() {
				Expect(config.MasterDatabaseURL()).To(Equal("postgres://foobar:9Gp0efB5VYBdeOu-TnbTb5VqjnsLFXw7rUV55SidDk8@localhost:15432/sandbox"))
			})
		})

		Context("for the standby cluster", func() {
			It("has the configured server version", func() {
				Expect(config.Standby.Version).ToNot(BeEmpty())
				Expect(config.Standby.Version).To(Equal("11"))
			})

			It("has the configured standby cluster name", func() {
				Expect(config.Standby.ClusterName).ToNot(BeEmpty())
				Expect(config.Standby.ClusterName).To(Equal("main"))
			})

			It("has the database URL", func() {
				Expect(config.StandbyDatabaseURL()).To(Equal("postgres://foobar:9Gp0efB5VYBdeOu-TnbTb5VqjnsLFXw7rUV55SidDk8@localhost:16432/sandbox"))
			})
		})

		Context("for pgbackrest", func() {
			It("has the configured stanza", func() {
				Expect(config.PgBackRest.Stanza).ToNot(BeEmpty())
				Expect(config.PgBackRest.Stanza).To(Equal("pitr"))
			})
		})

		Context("for minio", func() {
			It("has a local port", func() {
				Expect(config.Minio.LocalPort).ToNot(BeNil())
				Expect(config.Minio.LocalPort).To(BeNumerically(">", 0))
			})

			It("has an access key", func() {
				Expect(config.Minio.AccessKey).ToNot(BeEmpty())
			})

			It("has an secret key", func() {
				Expect(config.Minio.SecretKey).ToNot(BeEmpty())
			})
		})
	})
})
