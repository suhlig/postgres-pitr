package postgres_pitr_test

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"strings"

	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	DB struct {
		Host string
		Port int
		Name string
		User string
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

func (cfg Config) URL() (string, error) {
	password, err := cfg.Password()

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.DB.User, password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name), nil
}

var _ = Describe("Backup", func() {
	var db *sql.DB
	var err error

	BeforeEach(func() {
		cfg, err := NewConfig("../config.yml")
		Expect(err).NotTo(HaveOccurred())

		url, err := cfg.URL()
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
})
