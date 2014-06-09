## pgpin

pgpin is a toy clone of [Heroku Dataclips](https://dataclips.heroku.com),
aiming to be an open-source example of building
database-backed services in Go.

pgpin is an in-progress experiment. We hope to learn:

* What an idiomatic database-backed service in Go looks
  like.

* How viable it is to develop such services in Go, vs. e.g.
  the Sinatra/Sequel stack.

If the experiment is successful, we should end up with some
nice artifacts:

* A non-trivial example service in Go.

* A list of key functionalities for such services and
  example implementation snippets in the app.

* Documentation on development workflow around the app,
  including setting up a dev environment, running tests,
  deploying, etc.

### Status

* [FEATURES.md](FEATURES.md): things we've done and
  therefore have useful examples of.
* [TODO.md](TODO.md): things we'd like to do or investigate.

### Developing

A [Vagrant](http://www.vagrantup.com/) development
environment is provided. Install a recent version of Vagrant
and Virtualbox and run:

```console
$ vagrant up
$ vagrant ssh
$ cd src/github.com/mmcgrana/pgpin/pgpin-api
```

To start a development version of app:

```console
$ cat db/* | psql $DATABASE_URL
$ godep go install
$ goreman start
```

An environment variable is provided to make testing with
curl easy:

```console
$ curl $PGPIN_API_URL/status
```

To apply code changes:

```console
$ godep go install
$ goreman start
```

To run tests:

```console
$ cat db/* | psql $DATABASE_URL
$ godep go test
```

By default logs are silenced during tests. Turn them on
with:

```console
$ TEST_LOGS=true godep go test
```
