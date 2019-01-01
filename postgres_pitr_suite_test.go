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
	Expect(hosts).ToNot(BeEmpty())
	postgresHost = *hosts["master"]
})

func TestPostgresPitr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PostgreSQL PITR Suite")
}
