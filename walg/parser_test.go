package walg_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/suhlig/postgres-pitr/walg"
)

var _ = Describe("list-backups output parser", func() {
	Context("some backups exist", func() {
		stdout := `Path:  foobar/
name                          last_modified        wal_segment_backup_start
base_000000010000000000000003 2019-01-11T12:04:40Z 000000010000000000000003
base_000000010000000000000005 2019-01-11T15:27:46Z 000000010000000000000005
base_000000010000000000000007 2019-01-11T15:30:13Z 000000010000000000000007
base_000000010000000000000009 2019-01-11T15:30:27Z 000000010000000000000009
base_00000001000000000000000B 2019-01-11T15:34:20Z 00000001000000000000000B
`
		var info *walg.Info
		var err error

		JustBeforeEach(func() {
			info, err = walg.ParseListOutput(stdout)
			Expect(err).NotTo(HaveOccurred())
		})

		It("has the path", func() {
			Expect(info.Path).NotTo(BeEmpty())
			Expect(info.Path).To(Equal("foobar/"))
		})

		Context("all backups", func() {
			var backups []walg.Backup

			JustBeforeEach(func() {
				backups = info.Backups
			})

			It("has the expected number of backups", func() {
				Expect(backups).NotTo(BeEmpty())
				Expect(len(backups)).To(Equal(5))
			})

			Context("the first backup", func() {
				var firstBackup walg.Backup

				JustBeforeEach(func() {
					firstBackup = backups[0]
				})

				It("has the expected name", func() {
					Expect(firstBackup.Name).To(Equal("base_000000010000000000000003"))
				})

				It("has the expected timestamp", func() {
					Expect(firstBackup.LastModified).To(Equal(time.Date(2019, 1, 11, 12, 4, 40, 0, time.UTC)))
				})

				It("has the WAL segment with the expected backup start", func() {
					Expect(firstBackup.WalSegmentBackupStart).To(Equal("000000010000000000000003"))
				})
			})

			Context("the last backup", func() {
				var lastBackup walg.Backup

				JustBeforeEach(func() {
					lastBackup = backups[len(backups)-1]
				})

				It("has the expected name", func() {
					Expect(lastBackup.Name).To(Equal("base_00000001000000000000000B"))
				})

				It("has the expected timestamp", func() {
					Expect(lastBackup.LastModified).To(Equal(time.Date(2019, 1, 11, 15, 34, 20, 0, time.UTC)))
				})

				It("has the WAL segment with the expected backup start", func() {
					Expect(lastBackup.WalSegmentBackupStart).To(Equal("00000001000000000000000B"))
				})
			})
		})
	})
})
