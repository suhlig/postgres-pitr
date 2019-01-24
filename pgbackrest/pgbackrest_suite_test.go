package pgbackrest_test

import (
	"math/rand"
	"testing"

	"github.com/mikkeloscar/sshconfig"
	"github.com/suhlig/postgres-pitr/vagrant"

	. "github.com/onsi/ginkgo"
	config "github.com/onsi/ginkgo/config"
	. "github.com/onsi/gomega"
)

var masterHost sshconfig.SSHHost
var standbyHost sshconfig.SSHHost

var _ = BeforeSuite(func() {
	hosts, err := vagrant.Hosts()
	Expect(err).NotTo(HaveOccurred())
	Expect(hosts).ToNot(BeEmpty())

	masterHost = *hosts["master"]
	standbyHost = *hosts["standby"]

	rand.Seed(config.GinkgoConfig.RandomSeed)
})

func TestPostgresPitr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PgBackRest Suite")
}
