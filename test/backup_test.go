package postgres_pitr_test

import (
	"database/sql"

	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Backup", func() {
	var db *sql.DB
	var err error

	BeforeEach(func() {
		db, err = sql.Open("postgres", "postgres://foobar:DAyOx5UqeJtl2strwQyp@localhost:15432/sandbox")
		Expect(err).NotTo(HaveOccurred())
	})

	It("has the expected server version", func() {
		var rows *sql.Rows
		rows, err = db.Query("SHOW server_version_num;")
		defer rows.Close()
		Expect(err).NotTo(HaveOccurred())
		Expect(rows.Next()).To(BeTrue())

		var version int
		err = rows.Scan(&version)
		Expect(err).NotTo(HaveOccurred())
		Expect(version).To(BeNumerically(">=", 110000))
	})
})
