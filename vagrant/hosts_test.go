package vagrant_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/suhlig/postgres-pitr/sshrunner"
)

var _ = Describe("Vagrant Hosts", func() {
	var ssh *sshrunner.Runner
	var err error

	BeforeEach(func() {
		ssh, err = ssh.New(*hosts[0])
		Expect(err).NotTo(HaveOccurred())
	})

	It("can connect using SSH", func() {
		stdout, stderr, err := ssh.Run("id")
		Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
		Expect(stdout).To(ContainSubstring("vagrant"))
	})

	It("can run commands with args via SSH", func() {
		stdout, stderr, err := ssh.Run("ls -l")
		Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
		Expect(stdout).To(ContainSubstring("total"))
	})
})
