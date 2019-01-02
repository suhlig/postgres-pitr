package sshrunner_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	ssh "github.com/suhlig/postgres-pitr/sshrunner"
)

var _ = Describe("SSH Runner", func() {
	var ssh *ssh.Runner
	var err error

	BeforeEach(func() {
		Expect(len(hosts)).To(BeNumerically(">=", 1), "Expect at least one host, but found %d", len(hosts))

		host := hosts["master"]
		Expect(host).ToNot(BeNil())

		ssh, err = ssh.New(*host)
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

	It("provides error output", func() {
		stdout, stderr, err := ssh.Run("psql")
		Expect(err).To(HaveOccurred())
		Expect(stdout).To(BeEmpty())
		Expect(stderr).ToNot(BeEmpty())
	})
})
