## pgpin

pgpin is a toy clone of [Heroku Dataclips](https://dataclips.heroku.com),
written as an example Go app with an API, CLI, and web interface.

Like Dataclips, pgpin lets you construct SQL queries, continuously
refreshes query results in the background, stores queries for later
review, shows most recent query view results on the web, and allows
sharing of results with colleagues using secret URLs. It's like gist
for SQL queries.

For more info on the Go example code within pgpin, please see the
blog post [An example Go app: pgpin](https://mmcgrana.github.io/posts/2014-06-example-go-app-pgpin.html).

For pgpin usage and for deploying and developing the service, please
see below.

* [Usage](#usage)
* [Deploying](#deploying)
* [Developing](#developing)

### Usage

We'll use the CLI interface to show basic usage of pgpin. Similar
functionality is available on the web interface and the [API](api-docs).

To start using the CLI, configure it with a `PGPIN_API_URL`:

```console
$ PGPIN_DEPLOY=you
$ PGPIN_AUTH=heroku config:get AUTH -a pgpin-api-$PGPIN_DEPLOY
$ export PGPIN_API_URL="https://$PGPIN_AUTH@pgpin-api-$PGPIN_DEPLOY.herokuapp.com"
```

Verify we can connect to the API:

```console
$ pgpin api-status
ok
```

Get overview help:

```console
$ pgpin help
Usage: pgpin <command> [options] [arguments]

Commands:
  pins        List pins
  pin-create  Create new pin
  pin-delete  Delete a pin
  pin-show    Show pin details
  dbs         List databases
  db-add      Add database to registry
  db-remove   Remove database from registry
  db-show     Show database details
  api-status  Check API status
  help        Show help
```

To start, we'll need to add a database against which we can make
pins:

```console
$ pgpin db-add --name test-database --url "postgres://user:pass@host:1234/database"
33d03fe9ac28
```

See that it's indeed registered, and check the registered details:

```console
$ pgpin dbs
33d03fe9ac28 test-database

$ pgpin db-show test-database
Id:         33d03fe9ac28
Name:       test-database
Added At:   2014-06-01T17:55:03Z
Url:        postgres://user:pass@host:1234/database
```

Now let's create a pin against this database:

```console
$ pgpin pin-create --name test-pin --db test-database --query "select count(*) from users"
0b8b725f5750
```

The pin is persisted by the pgpin service:

```console
$ pgpin pins
0b8b725f5750 test-pin
```

See the details of your pin, including query results:

```console
$ pgpin pin-show test-pin
Id:                0b8b725f5750
Name:              test-pin
Db Id:             33d03fe9ac28
Created At:        2014-06-01T17:55:36Z
Query Started At:  2014-06-01T17:55:36Z
Query Finished At: 2014-06-01T18:08:48Z

Query:
select count(*) from users

Results:
 (?column?) |
------------+
 127        |
```

We can remove our test pin and database with:

```console
$ pgpin pin-delete test-pin

$ pgpin db-remove test-database
```

### Deploying

Deploy an instance of `pgping-api` and `pgpin-web` to Heroku with:

```console
$ export DEPLOY=you
$ bin/deploy-setup
```

This will give apps e.g. `pgping-api-you` and `pgpin-web-you`.

Tear these apps down with:

```console
$ bin/deploy-teardown
```

### Developing

A [Vagrant](http://www.vagrantup.com/) development environment is
provided. Install a recent version of Vagrant and Virtualbox and
run:

```console
$ vagrant up
$ vagrant ssh
$ cd src/github.com/mmcgrana/pgpin/pgpin-api
```

To start a development version of app:

```console
$ cat db/* | psql $DATABASE_URL
$ godep go install
$ goreman start
```

To apply changes:

```console
$ godep go install
$ goreman start
```

To run tests:

```console
$ cat db/* | psql $DATABASE_URL
$ godep go test
```

By default logs are silenced during tests. Turn them on
with:

```console
$ export TEST_LOGS=true
```
