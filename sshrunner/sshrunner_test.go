package sshrunner_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/suhlig/postgres-pitr/sshrunner"
)

var _ = Describe("SSH Runner", func() {
	var ssh *sshrunner.Runner
	var err error

	BeforeEach(func() {
		Expect(len(hosts)).To(BeNumerically("==", 1), "Expect exactly one host, but found %d", len(hosts))
		ssh, err = ssh.New(*hosts[0])
		Expect(err).NotTo(HaveOccurred())
	})

	It("can run a command", func() {
		stdout, stderr, err := ssh.Run("id")
		Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
		Expect(stdout).To(ContainSubstring("vagrant"))
	})

	It("can run a command with args", func() {
		stdout, stderr, err := ssh.Run("ls -l")
		Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
		Expect(stdout).To(ContainSubstring("total"))
	})
})
