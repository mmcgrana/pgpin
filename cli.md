## CLI

Get overview help:

```console
$ datapins
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
 Id
-------------------------------------+
4c15dbdc-4f8f-11e2-80dc-1040f386e726 |
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
