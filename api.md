## API

Datapins offers a JSON API at datapins-api-production.herokuapp.com.

The API authenticates users with Heroku API tokens or OAuth keys. To try the API with curl, set up your ~/.netrc:

```console
$ export HOST=datapins-api-production.herokuapp.com
$ cat ~/.netrc | grep -A 2 "machine api.heroku.com" | sed "s/api.heroku.com/$HOST/" >> ~/.netrc
```

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
    "id": "57238976-4f84-11e2-80d7-1040f386e726",
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
  "id": "57238976-4f84-11e2-80d7-1040f386e726",
  "resource_id": "resource274@heroku.com",
  "name": "posts count",
  "sql", "select count(*) from posts",
  "created_at": "2012/05/24 06:02:31 -0000",
  "user_id": "user248@heroku.com",
  "results_json": null,
  "results_at": null
}
```

Get a pin:

```console
$ export ID=b91376ba-4f83-11e2-8025-1040f386e726
$ curl -ns https://$HOST/v1/pins/$ID
{
  "id": "57238976-4f84-11e2-80d7-1040f386e726",
  "resource_id": "resource274@heroku.com",
  "name": "posts count",
  "sql", "select count(*) from posts",
  "created_at": "2012/05/24 06:02:31 -0000",
  "user_id": "user248@heroku.com",
  "results_json": "...",
  "results_at": "2012/05/24 06:02:33 -0000"
}
```

Destroy a pin:

```console
$ export ID=b91376ba4f83-11e2-8025-1040f386e726
$ curl -ns -X DELETE https://$HOST/v1/pins/$ID
{
  "id": "57238976-4f84-11e2-80d7-1040f386e726",
  "resource_id": "resource274@heroku.com",
  "name": "posts count",
  "sql", "select count(*) from posts",
  "created_at": "2012/05/24 06:02:31 -0000",
  "user_id": "user248@heroku.com",
  "results_json": "...",
  "results_at": "2012/05/24 06:02:33 -0000"
}
```

Check the status of the Datapins service:

```console
$ curl -x https://$HOST/v1/status
{
  "message": "ok"
}
