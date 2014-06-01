# Datapins API

## Deployment

```console
$ export DEPLOY=`whoami`
$ bin/deploy-setup
$ bin/deploy-teardown
```

You can start a local deploy shimmed into your development deploy with:

```console
$ heroku config -a $APP -s > .env
$ go get
$ foreman start
```
