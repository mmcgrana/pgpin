## pgpin

`pgpin` is an open-source example app showing how to build
a database-backed service in Go.

The example is a clone of [Heroku Dataclips](https://dataclips.heroku.com),
basically a "pastebin for SQL queries".

Work on `pgpin` is in progress. The current version of the app has
some but not all of the user-facing features and robustness
properties we eventually want.

In working through this example app we hope to learn:

* How to build database-backed services in idiomatic Go, with an
  eye towards robustness and operability.

* How developing such services in Go compares to other
  stacks, such as Ruby/Sinatra/Sequel.

The scope of finished and outstanding work is described in,
respectively:

* [FEATURES.md](FEATURES.md)
* [TODO.md](TODO.md)

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

### License

See [LICENSE.md](LICENSE.md).

### References

Some other projects we've looked to in writing the `pgpin`
example app:

* [Gddo server for godoc.org](https://github.com/golang/gddo)
* [Interagent HTTP API design guide](https://github.com/interagent/http-api-design)
* [Pliny Ruby toolkit](https://github.com/interagent/pliny)

We are interested in collecting other examples of non-trivial,
open-source, database-backed (Postgres or otherwise) services
written in Go. Please send suggestions to the
[author](https://twitter.com/mmcgrana).
