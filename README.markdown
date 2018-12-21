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

## Iterate

* Run tests when they changed:

  ```sh
  $ ginkgo watch -v -r
  ```

* Provision using Ansible when a playbook file has changed:

  ```sh
  $ fswatch -r ansible/**/* | xargs -I {} vagrant provision
  ```

# References

* [anishnath](https://github.com/anishnath/postgres) describes a manual approach
* The [Apcelent Tech Blog](https://blog.apcelent.com/using-ansible-to-set-up-postgresql.html) lists a few Ansible roles
* Federico Campoli's [Ansible roles](https://github.com/the4thdoctor/dynamic_duo/blob/04_pgbackrest/roles/rollback/tasks/rollback_ssh.yml) are interesting

# TODO

* Add a VM as [Dedicated Repository Host](https://pgbackrest.org/user-guide.html#repo-host)
* Fix poor error handling in the controllers
* Test more than the happy path
