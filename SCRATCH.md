### Usage

We'll use the CLI interface to show basic usage of pgpin.
Similar functionality is available on the web interface and
the [API](api-docs).

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

To start, we'll need to add a database against which we can
make pins:

```console
$ pgpin db-add --name test-database --url "postgres://user:pass@host:1234/database"
33d03fe9ac28
```

See that it's indeed registered, and check the registered
details:

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

Deploy an instance of `pgping-api` and `pgpin-web` to Heroku
with:

```console
$ export DEPLOY=you
$ bin/deploy-setup
```

This will give apps e.g. `pgping-api-you` and
`pgpin-web-you`.

Tear these apps down with:

```console
$ bin/deploy-teardown
```

Check API status:

```console
$ curl -i -X GET $PGPIN_API_URL/status
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Mon, 02 Jun 2014 17:02:13 GMT
Content-Length: 22

{
  "message": "ok"
}
```

Add a database:

```console
$ cat > /tmp/db-input.json <<EOF
{
  "name": "pins",
  "url": "postgres://postgres:secret@127.0.0.1:5432/pgpin-development"
}
EOF

$ curl -i -X POST $PGPIN_API_URL/dbs -H "Content-Type: application/json" -d @/tmp/db-input.json
HTTP/1.1 201 Created
Content-Type: application/json; charset=utf-8
Date: Sat, 07 Jun 2014 00:36:10 GMT
Content-Length: 215

{
  "id": "5e9b470e218f",
  "name": "pins",
  "url": "postgres://postgres:secret@127.0.0.1:5432/pgpin-development",
  "added_at": "2014-06-07T00:36:10.826248539Z",
  "updated_at": "2014-06-07T00:36:10.826256369Z"
}
```

Review db details:

```console
$ curl -i -X GET $PGPIN_API_URL/dbs/5e9b470e218f
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Sat, 07 Jun 2014 00:36:35 GMT
Content-Length: 209

{
  "id": "5e9b470e218f",
  "name": "pins",
  "url": "postgres://postgres:secret@127.0.0.1:5432/pgpin-development",
  "added_at": "2014-06-07T00:36:10.826249Z",
  "updated_at": "2014-06-07T00:36:10.826256Z"
}

```

List dbs:

```console
$ curl -i -X GET $PGPIN_API_URL/dbs
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Sat, 07 Jun 2014 00:37:09 GMT
Content-Length: 57

[
  {
    "id": "5e9b470e218f",
    "name": "pins"
  }
]
```

Create a pin:

```console
$ cat > /tmp/pin-input.json <<EOF
{
  "db_id": "5e9b470e218f",
  "name": "pins-info",
  "query": "select name, db_id, query, created_at from pins"
}
EOF

$ curl -i -X POST $PGPIN_API_URL/pins -H "Content-Type: application/json" -d @/tmp/pin-input.json
HTTP/1.1 201 Created
Content-Type: application/json; charset=utf-8
Date: Sat, 07 Jun 2014 00:37:28 GMT
Content-Length: 371

{
  "id": "e8d53783e9f4",
  "name": "pins-info",
  "db_id": "5e9b470e218f",
  "query": "select name, db_id, query, created_at from pins",
  "created_at": "2014-06-07T00:37:28.796724843Z",
  "updated_at": "2014-06-07T00:37:28.796734351Z",
  "query_started_at": null,
  "query_finished_at": null,
  "results_fields": null,
  "results_rows": null,
  "results_error": null
}
```

See what results it got:

```console
$ curl -i -X GET $PGPIN_API_URL/pins/e8d53783e9f4
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Sat, 07 Jun 2014 00:37:44 GMT
Content-Length: 618

{
  "id": "e8d53783e9f4",
  "name": "pins-info",
  "db_id": "5e9b470e218f",
  "query": "select name, db_id, query, created_at from pins",
  "created_at": "2014-06-07T00:37:28.796725Z",
  "updated_at": "2014-06-07T00:37:28.895637Z",
  "query_started_at": "2014-06-07T00:37:28.861538Z",
  "query_finished_at": "2014-06-07T00:37:28.890658Z",
  "results_fields": [
    "name",
    "db_id",
    "query",
    "created_at"
  ],
  "results_rows": [
    [
      "pins-info",
      "5e9b470e218f",
      "select name, db_id, query, created_at from pins",
      "2014-06-07T00:37:28.796725Z"
    ]
  ],
  "results_error": null
}
```

List pins:

```console
$ curl -i -X GET $PGPIN_API_URL/pins
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Sat, 07 Jun 2014 00:38:15 GMT
Content-Length: 62

[
  {
    "id": "e8d53783e9f4",
    "name": "pins-info"
  }
]
```

Delete a pin:

```console
$ curl -i -X DELETE $PGPIN_API_URL/pins/e8d53783e9f4
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Mon, 02 Jun 2014 17:31:56 GMT
Content-Length: 419

{
  "id": "103f992f0662",
  "name": "pins count",
  "db_id": "bf8029eb5b07",
  "query": "select count(*) from pins",
  "created_at": "2014-06-02T17:04:31.996089Z",
  "query_started_at": "2014-06-02T17:04:32.114051Z",
  "query_finished_at": "2014-06-02T17:04:32.148756Z",
  "results_fields_json": "[\"count\"]",
  "results_rows_json": "[[1]]",
  "results_error": null,
  "deleted_at": "2014-06-02T17:31:56.288976924Z"
}
```
