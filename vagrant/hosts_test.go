package vagrant_test

import (
	"github.com/mikkeloscar/sshconfig"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/suhlig/postgres-pitr/vagrant"
)

var _ = Describe("Vagrant Hosts", func() {
	var hosts map[string]*sshconfig.SSHHost
	var err error

	BeforeEach(func() {
		hosts, err = vagrant.Hosts()
		Expect(err).NotTo(HaveOccurred())
	})

	It("has at least one host", func() {
		Expect(hosts).ToNot(BeEmpty())
	})

	It("retrieves a host by name", func() {
		Expect(hosts["postgres"]).ToNot(BeNil())
	})
})
