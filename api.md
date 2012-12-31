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
