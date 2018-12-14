# Spike on Postgres PITR Backup & Restore

# Development

Provision if the playbook has changed:

```sh
$ fswatch -r ansible/playbook.yml | xargs -I {} vagrant provision
```

# References

* [anishnath](https://github.com/anishnath/postgres)
* [Apcelent Tech Blog](https://blog.apcelent.com/using-ansible-to-set-up-postgresql.html)
