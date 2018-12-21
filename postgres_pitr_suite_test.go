package postgres_pitr_test

import (
	"testing"

	"github.com/mikkeloscar/sshconfig"
	"github.com/suhlig/postgres-pitr/vagrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var postgresHost sshconfig.SSHHost

var _ = BeforeSuite(func() {
	hosts, err := vagrant.Hosts()
	Expect(err).NotTo(HaveOccurred())
	Expect(len(hosts)).To(BeNumerically(">=", 1), "Expect exactly one host, but found %d", len(hosts))
	postgresHost = *hosts[1]
})

func TestPostgresPitr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PostgreSQL PITR Suite")
}
