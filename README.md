## pgpin

pgpin is a toy clone of [Heroku Dataclips](https://dataclips.heroku.com)
written to experiment with app development in Go.

Like Dataclips, pgpin let's you construct SQL queries, view the
results, and share the results as a web page using a secret URL.

Pgpin includes an API (`pgpin-api`), a web interface (`pgpin-web`),
and CLI (`pgpin`).

### CLI

Get overview help:

```console
$ datapins-cli
Usage: datapins-cli <command> [<args>]

Commands:
  resources  List available resources
  list       List datapins
  create     Create a new datapin
  show       Show dataping metadata
  destroy    Destroy a datapin
  status     Check service status

$ datapins-cli resources
 Id                       | Name                   | Attachments
--------------------------+------------------------+----------------------
resource1822@heroku.com   | laughing-loudly-2742   | shogun:green
...

$ datapins-cli list
 Id                                  | Name             
-------------------------------------+------------------
4c15dbdc-4f8f-11e2-80dc-1040f386e726 | cips count
...

$ datapins-cli create --resource "resource1822@heroku.com" --name "post count" --sql "select count(*) from posts"
Creating datapin... done
Id: 5ab73e4c-4f8f-11e2-92cd-1040f386e726

$ datapins-cli show --id 5ab73e4c-4f8f-11e2-92cd-1040f386e726
Id:          5ab73e4c-4f8f-11e2-92cd-1040f386e726
Resource Id: resource1822@heroku.com
Name:        post count
Created At:  2012/05/24 06:02:31 -0000
Results At:  2012/05/24 06:02:33 -0000

Sql:
select count(*) from posts

Results:
 (?column?) |
------------+
 1          |
(1 row)

$ datapins-cli destroy --id 5ab73e4c-4f8f-11e2-92cd-1040f386e726
Destroying datapin... done

$ datapins-cli status
ok
```

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

