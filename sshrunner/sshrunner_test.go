package sshrunner_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/suhlig/postgres-pitr/sshrunner"
)

var _ = Describe("SSH Runner", func() {
	var dbSSH *sshrunner.Runner
	var err error

	BeforeEach(func() {
		Expect(len(hosts)).To(BeNumerically(">=", 1), "Expect at lease one host, but found %d", len(hosts))

		host := *hosts[0]
		Expect(host.Host).To(ContainElement("postgres"))

		dbSSH, err = dbSSH.New(host)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can run a command", func() {
		stdout, stderr, err := dbSSH.Run("id")
		Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
		Expect(stdout).To(ContainSubstring("vagrant"))
	})

	It("can run a command with args", func() {
		stdout, stderr, err := dbSSH.Run("ls -l")
		Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
		Expect(stdout).To(ContainSubstring("total"))
	})
})
