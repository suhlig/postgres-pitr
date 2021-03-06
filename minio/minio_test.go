package minio_test

import (
	"crypto/tls"
	"fmt"
	"net/http"

	s3 "github.com/minio/minio-go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/suhlig/postgres-pitr/config"
)

var _ = Describe("Minio", func() {
	var config config.Config
	var err error
	var s3c *s3.Client

	BeforeEach(func() {
		config, err = config.FromFile("../config.yml")
		Expect(err).NotTo(HaveOccurred())

		s3c, err = s3.New(
			fmt.Sprintf("%s:%d", config.Minio.Host, config.Minio.Port),
			config.Minio.AccessKey,
			config.Minio.SecretKey,
			config.Minio.UseSSL,
		)
		Expect(err).ToNot(HaveOccurred())

		s3c.SetCustomTransport(&http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}})
	})

	Context("at least one bucket exists", func() {
		BeforeEach(func() {
			err = s3c.MakeBucket("system-test", "")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			s3c.RemoveBucket("system-test")
		})

		It("lists the bucket", func() {
			buckets, err := s3c.ListBuckets()

			Expect(err).ToNot(HaveOccurred())
			Expect(buckets).ToNot(BeEmpty())
		})
	})
})
