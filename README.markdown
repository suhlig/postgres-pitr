# Spike on Postgres PITR

This is basically an acceptance test implementing the [Point-in-Time Recovery
](https://pgbackrest.org/user-guide.html#pitr) chapter of the [pgBackRest User Guide
](https://pgbackrest.org/user-guide.html).

# TL;DR

```sh
$ git clone https://github.com/suhlig/postgres-pitr
$ cd postgres-pitr
$ scripts/setup
$ tmuxinator
```

When done, issue `tmuxinator stop local`, and the VM will be shut down, too.

# Development

This project requires Go >= v1.11 because we are using modules.

## Setup

```sh
$ scripts/setup
```

## Run Tests

```sh
$ bin/ginkgo -v -r
```

## Iterate

* Run tests when they changed:

  ```sh
  $ bin/ginkgo watch -v -r
  ```

* Provision using Ansible when a deployment-related file has changed or a new one was added:

  ```sh
  $ scripts/watch
  ```

# References

* [anishnath](https://github.com/anishnath/postgres) describes a manual approach
* The [Apcelent Tech Blog](https://blog.apcelent.com/using-ansible-to-set-up-postgresql.html) lists a few Ansible roles
* Federico Campoli's [Ansible roles](https://github.com/the4thdoctor/dynamic_duo/blob/04_pgbackrest/roles/rollback/tasks/rollback_ssh.yml) are interesting

# TODO

* Test encryption
* Add a VM as [Dedicated Repository Host](https://pgbackrest.org/user-guide.html#repo-host)
* Restore to a separate (read-only) host (a.k.a. [Hot Standby](https://pgbackrest.org/user-guide.html#replication/hot-standby)) and, using the restored cluster, verify that the restore works
* Fix poor error handling in the controllers
* Test more than the happy path
