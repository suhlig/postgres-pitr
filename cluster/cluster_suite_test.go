package cluster_test

import (
	"testing"

	"github.com/mikkeloscar/sshconfig"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cfg "github.com/suhlig/postgres-pitr/config"
	"github.com/suhlig/postgres-pitr/vagrant"
)

var (
	masterHost sshconfig.SSHHost
	config     cfg.Config
	err        error
)

var _ = BeforeSuite(func() {
	config, err = config.FromFile("../config.yml")
	Expect(err).NotTo(HaveOccurred())

	hosts, err := vagrant.Hosts()
	Expect(err).NotTo(HaveOccurred())
	Expect(hosts).ToNot(BeEmpty())

	masterHost = *hosts["master"]
})

func TestCluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cluster Suite")
}
