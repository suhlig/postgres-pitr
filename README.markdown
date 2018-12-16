# Spike on Postgres PITR

# TL;DR

```sh
$ go get github.com/suhlig/postgres-pitr
$ cd $GOPATH/src/github.com/suhlig/postgres-pitr
$ tmuxinator local
```

When done, issue `tmuxinator stop local`, and the VM will be shut down, too.

# Development

## Setup

```sh
$ scripts/setup
```

## Iterate

* Run tests when they changed:

  ```sh
  $ ginkgo watch
  ```

* Provision using Ansible when a playbook file has changed:

  ```sh
  $ fswatch -r ansible/**/* | xargs -I {} vagrant provision
  ```

# References

* [anishnath](https://github.com/anishnath/postgres) describes a manual approach
* The [Apcelent Tech Blog](https://blog.apcelent.com/using-ansible-to-set-up-postgresql.html) lists a few Ansible roles
* Federico Campoli's [Ansible roles](https://github.com/the4thdoctor/dynamic_duo/blob/04_pgbackrest/roles/rollback/tasks/rollback_ssh.yml) are interesting
