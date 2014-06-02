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

Create a pin:

```console
$ cat > /tmp/pin-input.json <<EOF
{
  "db_id": "bf8029eb5b07",
  "name": "pins-count",
  "query": "select count(*) from pins"
}
EOF

$ curl -i -X POST $PGPIN_API_URL/pins -H "Content-Type: application/json" -d @/tmp/pin-input.json
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Mon, 02 Jun 2014 17:04:32 GMT
Content-Length: 332

{
  "id": "103f992f0662",
  "name": "pins count",
  "db_id": "bf8029eb5b07",
  "query": "select count(*) from pins",
  "created_at": "2014-06-02T17:04:31.996089355Z",
  "query_started_at": null,
  "query_finished_at": null,
  "results_fields_json": null,
  "results_rows_json": null,
  "results_error": null
}
```

See what results it got:

```console
$ curl -i -X GET $PGPIN_API_URL/pins/103f992f0662
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Mon, 02 Jun 2014 17:29:20 GMT
Content-Length: 391

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
  "results_error": null
}
```

List pins:

```console
$ curl -i -X GET $PGPIN_API_URL/pins
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Mon, 02 Jun 2014 17:30:05 GMT
Content-Length: 63

[
  {
    "id": "103f992f0662",
    "name": "pins count"
  }
]
```

Delete a pin:

```console
$ curl -i -X DELETE $PGPIN_API_URL/pins/103f992f0662
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
