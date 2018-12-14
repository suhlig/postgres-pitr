package postgres_pitr_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPostgresPitr(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PostgresPitr Suite")
}
