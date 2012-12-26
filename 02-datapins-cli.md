## CLI

Get overview help:

```console
$ datapins
Usage: datapins <command> [<args>]

Commands:
  resources  List available resources
  list       List datapins
  create     Create a new datapin
  destroy    Destroy a datapin
  help       Get help for a command

$ datapins resources
 Id                       | Name                   | Attachments
--------------------------+------------------------+----------------------
resource1822@heroku.com   | laughing-loudly-2742   | shogun:green

$ datapins create --resource 
```

Create a new datapin
Creating datapins... done
https://datapins.heroku.com/

$ datapins:create 