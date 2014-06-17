## pgpin

`pgpin` is an open-source example app showing how to build
a database-backed service in Go.

The example is a clone of [Heroku Dataclips](https://dataclips.heroku.com),
basically a "pastebin for SQL queries".

In working through this example app we hope to learn:

* How to build database-backed services in idiomatic Go, with an
  eye towards robustness and operability.

* How developing such services in Go compares to other
  stacks, such as Ruby/Sinatra/Sequel.

Work on `pgpin` is ongoing. The current version of the app has some
but not all of the user-facing features and robustness properties we
eventually want. See [FEATURES.md](FEATURES.md) and
[TODO.md](TODO.md).

### Developing

A [Vagrant](http://www.vagrantup.com/) development
environment is provided. Install a recent version of Vagrant
and Virtualbox and run:

```console
$ vagrant up
$ vagrant ssh
```

To start a development version of app:

```console
$ cat migrations/* | psql $DATABASE_URL
$ godep go install
$ goreman start
```

An environment variable is provided to make testing with
curl easy:

```console
$ curl -i $PGPIN_URL/status
```

To apply code changes:

```console
$ godep go install
$ goreman start
```

We suggest a git pre-commit hook to verify adherence to Go coding
standards:

```console
$ cat > .git/hooks/pre-commit <<EOF
#!/bin/bash

gofmt -d **/*.go 2>&1 | tee /tmp/pgpin-pre-commit
if [ $(wc -l < /tmp/pgpin-pre-commit) -gt "0" ]; then
  exit 1
fi
go vet github.com/mmcgrana/pgpin || exit 1
errcheck -ignore 'fmt:a^' github.com/mmcgrana/pgpin || exit 1
EOF
$ chmod +x .git/hooks/pre-commit
```

### Testing

To run tests:

```console
$ cat migrations/* | psql $TEST_DATABASE_URL
$ godep go test
```

By default logs are silenced during tests. Turn them on
with:

```console
$ TEST_LOGS=true godep go test
```

### Scripting

To write and run one-off scripts that use the `pgpin` app code:

```console
$ cat > script/count_pins.go <<EOF
package main

import (
	"log"
	"../pgpin"
)

func main() {
    pgpin.PgStart()
    count, _ := pgpin.PgCount("SELECT count(*) from pins")
    log.Printf("pins.count total=%d", count)
}
EOF

$ godep go run script/count_pins.go
```

### Deployment

To an instance of `pgpin` to Heroku:

```console
$ export DEPLOY=... (e.g. "production", or your username)
$ bin/deploy-setup
$ curl -i https://pgpin-$DEPLOY.herokuapp.com/status
```

To tear it down:

```console
$ bin/deploy-teardown
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
