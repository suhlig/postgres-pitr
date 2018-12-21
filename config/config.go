package config

import (
	"fmt"
	"io/ioutil"

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
		Password    string
	}

	PgBackRest struct {
		Stanza string
	}

	Minio struct {
		Port      int
		AccessKey string `yaml:"access_key"`
		SecretKey string `yaml:"secret_key"`
	}
}

// New creates a new Config struct from the given path to the config file
func (cfg Config) FromFile(path string) (Config, error) {
	cfg = Config{}
	cfg.DB.Host = "localhost"
	cfg.DB.Port = 5432

	yamlFile, err := ioutil.ReadFile(path)

	if err != nil {
		return cfg, err
	}

	err = yaml.Unmarshal(yamlFile, &cfg)

	return cfg, err
}

// DatabaseURL returns the URL to access the database
func (cfg Config) DatabaseURL() (string, error) {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name), nil
}
