# Spike on Postgres PITR

# TL;DR

```sh
$ go get github.com/suhlig/postgres-pitr
$ cd $GOPATH/github.com/suhlig/postgres-pitr
$ tmuxinator local
```

When done, issue `tmuxinator stop local`, and the VM will be shut down, too.

# Development

Provision if the playbook has changed:

```sh
$ fswatch -r ansible/playbook.yml | xargs -I {} vagrant provision
```

# Tests

## Setup

```sh
$ go get \
     github.com/onsi/ginkgo/ginkgo \
     github.com/onsi/gomega \
     github.com/lib/pq \
     gopkg.in/yaml.v2 \
     github.com/mikkeloscar/sshconfig
```

## Iterate

```sh
$ ginkgo watch
```

# References

* [anishnath](https://github.com/anishnath/postgres) describes a manual approach
* The [Apcelent Tech Blog](https://blog.apcelent.com/using-ansible-to-set-up-postgresql.html) lists a few Ansible roles
* Federico Campoli's [Ansible roles](https://github.com/the4thdoctor/dynamic_duo/blob/04_pgbackrest/roles/rollback/tasks/rollback_ssh.yml) are interesting
