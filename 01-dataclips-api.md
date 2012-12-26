The Dataclips API is served from api.dataclips.heroku.com:

```console
$ export HOST=api.dataclips.heroku.com
```

Dataclips authenticates users via Heroku API tokens or OAuth keys. To use the Dataclips API with curl, copy your api.heroku.com netrc entry to api.dataclips.heroku.com:

```console
$ cat ~/.netrc | grep -A 2 "machine api.heroku.com" | sed 's/api.heroku.com/api.dataclips.heroku.com/' >> ~/.netrc
```

```
Create a clip
```console
$ curl -n -s https://api.dataclips.heroku.com/v1/clips
                                             /v1/
```
