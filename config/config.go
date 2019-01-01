package config

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Master struct {
		Version     string
		ClusterName string `yaml:"cluster_name"`
		Host        string
		Port        int
		LocalPort   int `yaml:"local_port"`
		Name        string
		User        string
		Password    string
	}

	PgBackRest struct {
		Stanza string
	}

	Minio struct {
		Host      string
		LocalPort int `yaml:"local_port"`
		Port      int
		AccessKey string `yaml:"access_key"`
		SecretKey string `yaml:"secret_key"`
	}
}

// FromFile creates a new Config struct from the given path to the config file
func (cfg Config) FromFile(path string) (Config, error) {
	cfg = Config{}
	cfg.Master.Port = 5432
	cfg.Minio.Port = 443

	yamlFile, err := ioutil.ReadFile(path)

	if err != nil {
		return cfg, err
	}

	err = yaml.Unmarshal(yamlFile, &cfg)

	return cfg, err
}

// DatabaseURL returns the URL to access the database
func (cfg Config) DatabaseURL() (string, error) {
	return fmt.Sprintf("postgres://%s:%s@localhost:%d/%s", cfg.Master.User, cfg.Master.Password, cfg.Master.LocalPort, cfg.Master.Name), nil
}
