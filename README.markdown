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
* The `systemd` template is based on [minio-service](https://github.com/minio/minio-service/blob/master/linux-systemd/README.md )

# TODO

* Merge cluster and pgbackrest controllers
  => move `stop + restore + start` as a single operation into the pbbackrest controller
* Try to get rid of port forwarding by connecting to postgres via the 192.168.*.* network instead. Needs a different allow statement in postgres config.
* Separate *database* config (incl. db name, user and password) from DB *cluster* config
* Optionally restore only a selected databases; see https://pgbackrest.org/user-guide.html#restore/option-db-include
* Do not rely on the `main` cluster, but create a separate one (`sudo pg_createcluster 9.4 demo` etc.)
* Run `pg_create_restore_point` through native DB driver instead of SSH
* Fix poor error handling in the controllers
* Test more than the happy path
