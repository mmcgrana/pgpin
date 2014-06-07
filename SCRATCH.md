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
