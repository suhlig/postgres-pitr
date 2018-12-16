package postgres_pitr_test

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var configFileName string

var _ = BeforeSuite(func() {
	sshConfigString, _, err := run("vagrant", "ssh-config")
	Expect(sshConfigString).To(ContainSubstring("default"))
	Expect(err).NotTo(HaveOccurred())

	tmpfile, err := ioutil.TempFile("", "vagrant-ssh-config")
	configFileName = tmpfile.Name()
	Expect(err).NotTo(HaveOccurred())

	_, err = tmpfile.Write([]byte(sshConfigString))
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
