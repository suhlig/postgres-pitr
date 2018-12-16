package postgres_pitr_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func run(args ...string) (string, string, error) {
	cmd := exec.Command(args[0], args[1:]...)

	var stderr string

	out, err := cmd.Output()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			stderr = string(exitError.Stderr)
		}
	}

	return string(out), stderr, err
}

var configFileName string

var _ = BeforeSuite(func() {
	stdout, stderr, err := run("vagrant", "ssh-config")
	Expect(err).NotTo(HaveOccurred(), "stderr was:\n%v\n\nstdout was:\n'%v\n\n'", stderr, stdout)
	Expect(stdout).To(ContainSubstring("default"))

	tmpfile, err := ioutil.TempFile("", "vagrant-ssh-config")
	configFileName = tmpfile.Name()
	Expect(err).NotTo(HaveOccurred())

	_, err = tmpfile.Write([]byte(stdout))
	Expect(err).NotTo(HaveOccurred())

	err = tmpfile.Close()
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	defer os.Remove(configFileName)
})

func TestPostgresPitr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PostgreSQL PITR Suite")
}
