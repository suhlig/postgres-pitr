package vagrant_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVagrant(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vagrant Suite")
}
