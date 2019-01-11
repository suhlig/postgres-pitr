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
		Name        string
		User        string
		Password    string
	}

	Standby struct {
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
		Host      string
		Port      int
		UseSSL    bool   `yaml:"use_ssl"`
		AccessKey string `yaml:"access_key"`
		SecretKey string `yaml:"secret_key"`
	}
}

// FromFile creates a new Config struct from the given path to the config file
func (cfg Config) FromFile(path string) (Config, error) {
	cfg.Master.Port = 5432
	cfg.Standby.Port = 5432
	cfg.Minio.Port = 443
	cfg.Minio.UseSSL = true

	yamlFile, err := ioutil.ReadFile(path)

	if err != nil {
		return cfg, err
	}

	err = yaml.Unmarshal(yamlFile, &cfg)

	return cfg, err
}

// MasterDatabaseURL returns the URL to access the database on the master node
func (cfg Config) MasterDatabaseURL() (string, error) {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.Master.User, cfg.Master.Password, cfg.Master.Host, cfg.Master.Port, cfg.Master.Name), nil
}

// StandbyDatabaseURL returns the URL to access the database on the standby node
func (cfg Config) StandbyDatabaseURL() (string, error) {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.Standby.User, cfg.Standby.Password, cfg.Standby.Host, cfg.Standby.Port, cfg.Standby.Name), nil
}

// BlobstoreURL returns the URL to access the database on the standby node
func (cfg Config) BlobstoreURL() (string, error) {
	if cfg.Minio.UseSSL {
		return fmt.Sprintf("https://%s:%d/", cfg.Minio.Host, cfg.Minio.Port), nil
	}

	return fmt.Sprintf("http://%s:%d/", cfg.Minio.Host, cfg.Minio.Port), nil
}
