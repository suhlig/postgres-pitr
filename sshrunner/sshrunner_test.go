package sshrunner_test

import (
	"strconv"
	"strings"

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

	It("can run multiple commands under a single session", func() {
		ts1s, _, err := ssh.Run("date +%%s%%N")
		Expect(err).ToNot(HaveOccurred())

		ts1, err := strconv.ParseInt(strings.TrimSuffix(ts1s, "\n"), 10, 64)
		Expect(err).ToNot(HaveOccurred())

		ts2s, _, err := ssh.Run("date +%%s%%N")
		Expect(err).ToNot(HaveOccurred())

		ts2, err := strconv.ParseInt(strings.TrimSuffix(ts2s, "\n"), 10, 64)
		Expect(err).ToNot(HaveOccurred())

		Expect(ts1).Should(BeNumerically("~", ts2, 1e7))
	})

	FMeasure("it should run multiple commands in the same SSH session", func(b Benchmarker) {
		runtime := b.Time("runtime", func() {
			_, _, err := ssh.Run("date +%%s%%N")
			Expect(err).ToNot(HaveOccurred())
		})

		Expect(runtime.Seconds()).Should(BeNumerically("<", 2.0), "SSH should be fast")
	}, 50)
})
