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
$ PGPIN_API_KEY=heroku config:get -a pgpin-api-$PGPIN_DEPLOY
$ export PGPIN_API_URL="https://$PGPIN_API_KEY@pgpin-api-$PGPIN_DEPLOY.herokuapp.com"
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
  pin-destroy Destroy a pin
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
Id:         0b8b725f5750
Name:       test-pin
Db Id:      33d03fe9ac28
Created At: 2014-06-01T17:55:36Z
Results At: 2014-06-01T18:08:48Z

Query:
select count(*) from users

Results:
 (?column?) |
------------+
 1          |
```

We can remove our test pin and database with:

```console
$ pgpin pin-rm test-pin
0b8b725f5750

$ pgpin db-rm test-database
33d03fe9ac28
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
```


-----

{
  "id": "be193c3048eb7c508abb9617493937c5",
  "resource_id": "resource274@heroku.com",
  "name": "posts count",
  "sql", "select count(*) from posts",
  "created_at": "2012/05/24 06:02:31 -0000",
  "user_id": "user248@heroku.com",
  "results_fields_json": "...",
  "results_rows_json": "...",
  "error_message": null
  "results_at": "2012/05/24 06:02:33 -0000"
}

## API

Datapins offers a JSON API at datapins-api-production.herokuapp.com.


### Authentication

The API authenticates users with Heroku API tokens or OAuth keys. To try the API with curl, set up your ~/.netrc:

```console
$ export HOST=datapins-api-production.herokuapp.com
$ cat ~/.netrc | grep -A 2 "machine api.heroku.com" | sed "s/api.heroku.com/$HOST/" >> ~/.netrc
```


### Endpoints

Get the Heroku resources against which you can create pins:

```console
$ curl -ns https://$HOST/v1/resources
[
  {
    "id": "resource132@heroku.com",
    "name": "boiling-fortress-9685",
    "attachements": [
      {
        "app_name": "shogun",
        "config_var": "HEROKU_POSTGRESQL_BLACK_URL"
      },
      ...
    ]
  },
  ...
]
```

Get pins:

```console
$ curl -ns https://$HOST/v1/pins
[
  {
    "id": "be193c3048eb7c508abb9617493937c5",
    "name": "posts count"
  },
  ...
]
```

Create a pin:

```console
$ cat > pin.js <<EOF
{
  "resource_id": "resource232@heroku.com",
  "name": "posts count",
  "sql": "select count(*) from posts"
}
EOF
$ curl -ns -X POST https://$HOST/v1/pins -H "Content-Type: application/json" -d @pin.js
{
  "id": "be193c3048eb7c508abb9617493937c5",
  "resource_id": "resource274@heroku.com",
  "name": "posts count",
  "sql", "select count(*) from posts",
  "user_id": "user248@heroku.com",
  "created_at": "2012/05/24 06:02:31 -0000",
  "results_fields_json": null,
  "results_rows_json": null,
  "error_message": null,
  "results_at": null
}
```

Get a pin:

```console
$ export ID=be193c3048eb7c508abb9617493937c5
$ curl -ns https://$HOST/v1/pins/$ID
{
  "id": "be193c3048eb7c508abb9617493937c5",
  "resource_id": "resource274@heroku.com",
  "name": "posts count",
  "sql", "select count(*) from posts",
  "created_at": "2012/05/24 06:02:31 -0000",
  "user_id": "user248@heroku.com",
  "results_fields_json": "...",
  "results_rows_json": "...",
  "error_message": null
  "results_at": "2012/05/24 06:02:33 -0000"
}
```

Destroy a pin:

```console
$ export ID=be193c3048eb7c508abb9617493937c5
$ curl -ns -X DELETE https://$HOST/v1/pins/$ID
{
  "id": "be193c3048eb7c508abb9617493937c5",
  "resource_id": "resource274@heroku.com",
  "name": "posts count",
  "sql", "select count(*) from posts",
  "created_at": "2012/05/24 06:02:31 -0000",
  "user_id": "user248@heroku.com",
  "results_fields_json": "...",
  "results_rows_json": "...",
  "error_message": null
  "results_at": "2012/05/24 06:02:33 -0000"
}
```

Check the status of the Datapins service:

```console
$ curl -x https://$HOST/v1/status
{
  "message": "ok"
}
```

All endpoints return 200 on success.


### Errors

Non-200 status codes indicate an error:

* 400: Invalid body
* 401: Authorization required
* 403: Invalid field
* 404: Not found
* 500: Internal server error or unanticipated user error

Bodies for all error responses are of the form:

```
{
  "message": "message text"
}
```
