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
		config, err = config.New("../config.yml")
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Configuration file contains an entry", func() {
		Context("for the cluster", func() {
			It("has the configured server version", func() {
				Expect(config.DB.Version).ToNot(BeEmpty())
				Expect(config.DB.Version).To(Equal("11"))
			})

			It("has the configured cluster name", func() {
				Expect(config.DB.ClusterName).ToNot(BeEmpty())
				Expect(config.DB.ClusterName).To(Equal("main"))

			})
		})

		Context("for pgbackrest", func() {
			It("has the configured stanza", func() {
				Expect(config.PgBackRest.Stanza).ToNot(BeEmpty())
				Expect(config.PgBackRest.Stanza).To(Equal("pitr"))
			})
		})
	})
})