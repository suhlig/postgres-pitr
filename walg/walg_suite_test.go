package walg_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"math/rand"

	"github.com/mikkeloscar/sshconfig"
	"github.com/suhlig/postgres-pitr/vagrant"

	config "github.com/onsi/ginkgo/config"
)

var masterHost sshconfig.SSHHost

var _ = BeforeSuite(func() {
	hosts, err := vagrant.Hosts()
	Expect(err).NotTo(HaveOccurred())
	Expect(hosts).ToNot(BeEmpty())

	masterHost = *hosts["master"]

	rand.Seed(config.GinkgoConfig.RandomSeed)
})

func TestWalg(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WAL-G Suite")
}
