package vagrant_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/suhlig/postgres-pitr/vagrant"
)

var _ = Describe("Vagrant Hosts", func() {
	It("has one host", func() {
		hosts, err := vagrant.Hosts()
		Expect(err).NotTo(HaveOccurred())
		Expect(len(hosts)).To(BeNumerically("==", 1), "Expect exactly one host, but found %d", len(hosts))
	})
})
