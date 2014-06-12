## Todo

### Docs

* Publish: Twitter
* Publish: blog
* Publish: golang-nuts

### API

* Test for timeout endpoint
* Test for worker panic handling
* Try other user query result types, including bytes
* Encrypt resource URL
* Encrypt other sensitive fields
* Multiple workers / concurrent workers
* Handle dangling reservations
* Operational metrics
* Handling extra input fields on JSON
* Exception reporting
* Request Ids on requests
* Request Ids passed through
* Review Go docker image at https://github.com/GoogleCloudPlatform/golang-docker/blob/master/base/Dockerfile
* Cross-check against pliny
* Cross-check against HTTP API design guide
* JSON schema
* Remove goji graceful stuff from worker output
* API versioning
* API authorization
* HTTP basic authentication
* Remove goji/bind dependency?
* Investigate validation libraries, https://github.com/pengux/check?
* Validate URL for well-formedness
* Validate URL for reachability at create-time
* Validate names are unique
* Capturing HTTP response status in logger
* Generic not found vs. e.g. db not found
* Review and catalog error ids
* Apply tool to check for missing error handles
* Hook standard go tools into git

### Client

* Implement proof of concept using Schematic

### CLI

* Implement proof of concept using hk/zk skeleton

### Web

* Implement proof of concept with Ember or React
