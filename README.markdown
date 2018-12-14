# Spike on Postgres PITR

# TL;DR

```sh
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
     github.com/lib/pq
```

## Iterate

```sh
$ ginkgo watch test
```

# References

* [anishnath](https://github.com/anishnath/postgres)
* [Apcelent Tech Blog](https://blog.apcelent.com/using-ansible-to-set-up-postgresql.html)
