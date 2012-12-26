Dataclips offers a JSON-over-HTTP API at api.dataclips.heroku.com.

The API authenticates users with Heroku API tokens or OAuth keys. To try the API with curl, set up your ~/.netrc:

```console
$ export HOST=api.dataclips.heroku.com
$ cat ~/.netrc | grep -A 2 "machine api.heroku.com" | sed "s/api.heroku.com/$HOST/" >> ~/.netrc
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
