Dataclips offers a JSON-over-HTTP API at api.dataclips.heroku.com.

It authenticates users with Heroku API tokens or OAuth keys. To try the API with curl, copy your api.heroku.com netrc entry to api.dataclips.heroku.com:

```console
$ cat ~/.netrc | grep -A 2 "machine api.heroku.com" | sed "s/api.heroku.com/api.dataclips.heroku.com/" >> ~/.netrc
```

Create a clip:

```console
$ cat > clip.js <<EOF
{
  "heroku_resource_id": "resource232@heroku.com",
   "sql": "select count(*) from posts"
}
EOF
$ curl -ns -X POST https://api.dataclips.heroku.com/v1/clips -H "Content-Type: application/json" -d @clip.js
```
