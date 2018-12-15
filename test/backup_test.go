package postgres_pitr_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	DB struct {
		Version     string
		ClusterName string `yaml:"cluster_name"`
		Host        string
		Port        int
		Name        string
		User        string
	}

	PgBackRest struct {
		Stanza string
	}
}

type Info struct {
	Name   string
	Status struct {
		Code    int
		Message string
	}
}

func NewConfig(path string) (Config, error) {
	cfg := Config{}
	cfg.DB.Host = "localhost"
	cfg.DB.Port = 5432

	yamlFile, err := ioutil.ReadFile(path)

	if err != nil {
		return cfg, err
	}

	err = yaml.Unmarshal(yamlFile, &cfg)

	return cfg, err
}

func (cfg Config) Password() (string, error) {
	password, err := ioutil.ReadFile("../ansible/.postgres-password")

	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(string(password), "\n"), nil
}

func (cfg Config) DatabaseURL() (string, error) {
	password, err := cfg.Password()

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.DB.User, password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name), nil
}

func runCommand(args ...string) (string, string, error) {
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

func vagrantSSH(command string, args ...interface{}) (string, string, error) {
	commands := []string{"vagrant", "ssh", "-c", fmt.Sprintf(command, args...)}
	return runCommand(commands...)
}

var _ = Describe("pgBackRest", func() {
	var db *sql.DB
	var config Config
	var err error

	BeforeEach(func() {
		config, err = NewConfig("../config.yml")
		Expect(err).NotTo(HaveOccurred())

		url, err := config.DatabaseURL()
		Expect(err).NotTo(HaveOccurred())

		db, err = sql.Open("postgres", url)
		Expect(err).NotTo(HaveOccurred())
	})

	It("can connect", func() {
		err = db.Ping()
		Expect(err).NotTo(HaveOccurred())
	})

	It("has the expected server version", func() {
		var version int
		err = db.QueryRow("SHOW server_version_num;").Scan(&version)
		Expect(err).NotTo(HaveOccurred())
		Expect(version).To(BeNumerically(">=", 110000))
	})

	Context("config file", func() {
		Context("for the cluster", func() {
			It("has the configured server version", func() {
				Expect(config.DB.Version).To(Equal("11"))
			})

			It("has the configured cluster name", func() {
				Expect(config.DB.ClusterName).To(Equal("main"))
			})
		})

		Context("for pgbackrest", func() {
			It("has the configured stanza", func() {
				Expect(config.PgBackRest.Stanza).To(Equal("pitr"))
			})
		})
	})

	It("connects using SSH", func() {
		stdout, stderr, err := vagrantSSH("id")
		Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
		Expect(stdout).To(ContainSubstring("vagrant"))
	})

	It("can run commands with args via SSH", func() {
		stdout, stderr, err := vagrantSSH("ls -l")
		Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
		Expect(stdout).To(ContainSubstring("total"))
	})


	Context("A backup exists", func() {
		BeforeEach(func() {
			stdout, stderr, err := vagrantSSH("sudo -u postgres pgbackrest --stanza=%s backup", config.PgBackRest.Stanza)
			Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
		})

		It("lists the backup", func() {
			stdout, stderr, err := vagrantSSH("sudo -u postgres pgbackrest info --stanza=%s --output=json", config.PgBackRest.Stanza)
			Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)

			infos := make([]Info, 0)
			err = json.Unmarshal([]byte(stdout), &infos)
			Expect(err).ToNot(HaveOccurred(), "stdout was: '%v'", stdout)

			Expect(infos).To(HaveLen(1))
			info := infos[0]
			Expect(info.Name).To(Equal(config.PgBackRest.Stanza))
			Expect(info.Status.Code).To(Equal(0), "Message is %s", info.Status.Message)
			Expect(info.Status.Message).To(Equal("ok"))
		})

		When("the pg_control file is lost", func() {
			XIt("successfully restores the cluster", func() {

			})
		})

		When("important data was deleted", func() {
			XIt("can be restored to the state at the given point in time", func() {
				// TODO https://pgbackrest.org/user-guide.html#pitr
			})
		})

		/*
		   // Verify cluster status:
		   sudo pg_ctlcluster 11 main status

		   // Stop the main cluster and delete the pg_control file
		   sudo pg_ctlcluster 9.4 main stop
		   sudo -u postgres rm /var/lib/postgresql/9.4/main/global/pg_control

		   // Verify that starting the cluster without this important file will result in an error
		   sudo pg_ctlcluster 9.4 main start

		   // ensure cluster is stopped
		   sudo pg_ctlcluster 11 main status

		   // remove all files from the PostgreSQL data directory
		   sudo -u postgres find /var/lib/postgresql/9.4/main -mindepth 1 -delete

		   // Restore the main cluster and start PostgreSQL
		   sudo -u postgres pgbackrest --stanza=pitr restore
		   sudo pg_ctlcluster 9.4 main start

		   // Verify that the cluster is up
		*/
	})
})
