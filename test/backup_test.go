package postgres_pitr_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	_ "github.com/lib/pq"
	"github.com/mikkeloscar/sshconfig"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"
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

func run(args ...string) (string, string, error) {
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

type VagrantSSH struct {
	Host sshconfig.SSHHost
}

func NewVagrantSSH(host sshconfig.SSHHost) (*VagrantSSH, error) {
	vagrant := &VagrantSSH{Host: host}
	return vagrant, nil
}

func (vagrant *VagrantSSH) Run(command string, args ...interface{}) (string, string, error) {
	privateKey, err := ioutil.ReadFile(vagrant.Host.IdentityFile)

	if err != nil {
		return "", "", err
	}

	key, err := ssh.ParsePrivateKey(privateKey)
	if err != nil {
		return "", "", err
	}

	config := &ssh.ClientConfig{
		User:            vagrant.Host.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
	}

	addr := fmt.Sprintf("%s:%d", vagrant.Host.HostName, vagrant.Host.Port)
	client, err := ssh.Dial("tcp", addr, config)

	if err != nil {
		return "", "", err
	}

	session, err := client.NewSession()

	if err != nil {
		return "", "", err
	}

	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	err = session.Run(fmt.Sprintf(command, args...))

	return stdoutBuf.String(), stderrBuf.String(), err
}

var _ = Describe("pgBackRest", func() {
	var db *sql.DB
	var config Config
	var vagrant *VagrantSSH
	var err error

	BeforeEach(func() {
		config, err = NewConfig("../config.yml")
		Expect(err).NotTo(HaveOccurred())

		url, err := config.DatabaseURL()
		Expect(err).NotTo(HaveOccurred())

		// TODO START Move this to a proper type
		sshConfigString, _, err := run("vagrant", "ssh-config")
		Expect(sshConfigString).To(ContainSubstring("default"))
		Expect(err).NotTo(HaveOccurred())

		tmpfile, err := ioutil.TempFile("", "vagrant-ssh-config")
		Expect(err).NotTo(HaveOccurred())
		defer os.Remove(tmpfile.Name())

		_, err = tmpfile.Write([]byte(sshConfigString))
		Expect(err).NotTo(HaveOccurred())
		err = tmpfile.Close()
		Expect(err).NotTo(HaveOccurred())

		hosts, err := sshconfig.ParseSSHConfig(tmpfile.Name())
		Expect(len(hosts)).To(BeNumerically("==", 1), "Require exactly one host, but found %d", len(hosts))

		vagrant, err = NewVagrantSSH(*hosts[0])
		// TODO END Move this to a proper type

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
		stdout, stderr, err := vagrant.Run("id")
		Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
		Expect(stdout).To(ContainSubstring("vagrant"))
	})

	It("can run commands with args via SSH", func() {
		stdout, stderr, err := vagrant.Run("ls -l")
		Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
		Expect(stdout).To(ContainSubstring("total"))
	})

	Context("A backup exists", func() {
		BeforeEach(func() {
			stdout, stderr, err := vagrant.Run("sudo -u postgres pgbackrest --stanza=%s backup", config.PgBackRest.Stanza)
			Expect(err).ToNot(HaveOccurred(), "stderr was: '%v', stdout was: '%v'", stderr, stdout)
		})

		It("lists the backup", func() {
			stdout, stderr, err := vagrant.Run("sudo -u postgres pgbackrest info --stanza=%s --output=json", config.PgBackRest.Stanza)
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

		XWhen("the pg_control file is lost", func() {
			It("successfully restores the cluster", func() {

			})
		})

		XWhen("important data was deleted", func() {
			It("can be restored to the state at the given point in time", func() {
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
