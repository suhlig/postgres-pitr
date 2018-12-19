package sshrunner_test

import (
	"testing"

	"github.com/mikkeloscar/sshconfig"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/suhlig/postgres-pitr/vagrant"
)

var hosts []*sshconfig.SSHHost

var _ = BeforeSuite(func() {
	var err error
	hosts, err = vagrant.Hosts()
	Expect(err).NotTo(HaveOccurred())
})

func TestSshrunner(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sshrunner Suite")
}
