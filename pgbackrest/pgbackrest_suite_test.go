package pgbackrest_test

import (
	"math/rand"
	"strings"
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

func randomName() string {
	chars := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	length := 16
	var builder strings.Builder

	for i := 0; i < length; i++ {
		builder.WriteRune(chars[rand.Intn(len(chars))])
	}

	return builder.String()
}
