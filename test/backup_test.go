package postgres_pitr_test

import (
	"database/sql"
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

func RunCommand(args ...string) (string, string, error) {
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

func RunVagrantSSHCommand(args ...string) (string, string, error) {
	baseCommands := []string{"vagrant", "ssh", "-c"}
	return RunCommand(append(baseCommands, args...)...)
}

var _ = Describe("Backup", func() {
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
		})
	})
	It("connects using SSH", func() {
		stdout, stderr, err := RunVagrantSSHCommand("id")
		Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
		Expect(stdout).To(ContainSubstring("vagrant"))
	})

})
