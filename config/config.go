package config

import (
	"fmt"
	"io/ioutil"
	"strings"

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

func (_ Config) New(path string) (Config, error) {
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
	password, err := ioutil.ReadFile("ansible/.postgres-password")

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
