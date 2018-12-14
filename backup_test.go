package postgres_pitr_test

import (
	"database/sql"
	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Backup", func() {
	It("connects to the DB", func() {
		db, err := sql.Open("postgres", "postgres://foobar:DAyOx5UqeJtl2strwQyp@localhost:15432/sandbox")
		Expect(err).NotTo(HaveOccurred())

		_, err = db.Query("SELECT * FROM pg_catalog.pg_tables;")
		Expect(err).NotTo(HaveOccurred())
	})
})
