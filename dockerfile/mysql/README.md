# Introduction

The [MySQL](https://hub.docker.com/repository/docker/kryptondb/percona) image has been pushed into docker hub. The available versions are:

    kryptondb/percona (tag: 5.7)

Images are updated when new releases are published. 

# Environment Variables

## `MYSQL_ROOT_PASSWORD`

This variable specifies a password that will be set for the root superuser account.

**Notice**: Setting the MySQL root user password on the command line is insecure.

## `MYSQL_REPL_PASSWORD`

This variable specifies a replication password that will be set for the replication account, the default is `Repl_123`.

## `INIT_TOKUDB`

Set to `1` to allow the container to be started with enabled TOKUDB engine.

## `MYSQL_INITDB_SKIP_TZINFO`

Skip import time zone information into MySQL when the variable is set.

## `MYSQL_DATABASE`

This variable is optional. It allows you to specify the name of a database to be created on image startup. If a user/password was supplied (see below) then that user will be granted superuser access (corresponding to GRANT ALL) to this database.

## `MYSQL_USER`, `MYSQL_PASSWORD`

These variables are optional, used in conjunction to create a new user and set that user's password. This user will be granted superuser permissions (see above) for the database specified by the `MYSQL_DATABASE` variable. Both variables are required for a user to be created.

# Build Image

```
docker build -t mysql:v1.0 .
```
