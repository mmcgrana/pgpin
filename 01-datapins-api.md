## API

Datapins offers a JSON-over-HTTP API at datapins-api.herokuapp.com.

The API authenticates users with Heroku API tokens or OAuth keys. To try the API with curl, set up your ~/.netrc:

```console
$ export HOST=datapins-api.herokuapp.com
$ cat ~/.netrc | grep -A 2 "machine api.heroku.com" | sed "s/api.heroku.com/$HOST/" >> ~/.netrc
```

Get resources on which you can create clips:

```console
$ curl -ns https://$HOST/v1/resources
[
  {
    "id": "resource132@heroku.com",
    ...
  },
  ...
]
```

Get clips:

```console
$ curl -ns https://$HOST/v1/clips
[
  {
    "id": "57238976-4f84-11e2-80d7-1040f386e726",
    ...
  },
  {
    "id": "5e072982-4f84-11e2-99b2-1040f386e726",
    ...
  },
  ...
]
```

Create a clip:

```console
$ cat > clip.js <<EOF
{
  "heroku_resource_id": "resource232@heroku.com",
   "sql": "select count(*) from posts"
}
EOF
$ curl -ns -X POST https://$HOST/v1/clips -H "Content-Type: application/json" -d @clip.js
{
  "id": "b91376ba-4f83-11e2-8025-1040f386e726",
  ...
}
```

Get a clip:

```console
$ export ID=b91376ba-4f83-11e2-8025-1040f386e726
$ curl -ns https://$HOST/v1/clips/$ID
{
  "id": "b91376ba-4f83-11e2-8025-1040f386e726",
  ...
}
```

Update a clip:

```console
$ export ID=b91376ba-4f83-11e2-8025-1040f386e726
$ cat > clip.js <<EOF
{
  "sql": select count(id) from posts"
}
$ curl -ns -X PUT https://$HOST/v1/clips/$ID -H "Content-Type: application/json" -d @clips.js
{
  "id": "b91376ba-4f83-11e2-8025-1040f386e726"
}
```

Destroy a clip:

```console
$ export ID=b91376ba4f83-11e2-8025-1040f386e726
$ curl -ns -X DELETE https://$HOST/v1/clips/$ID
{
  "id": "b91376ba-4f83-11e2-8025-1040f386e726",
  ...
```
