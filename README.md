Mocker
======

A simple service to mock rest apis

Currently supports mocking GET requests only.  Mocker respects and enforces query string parameters.  Two requests to the same endpoint with different query string parameters may return different content as long as content is set for both sets of parameters.

Installation
------------

~~~~
git clone https://github.com/dmcgowan/mocker.git
cd mocker
export GOPATH=`pwd`
go get mocker
go install mocker
bin/mocker -port 8080
~~~~

API
---

### Create a new endpoint
`POST /mock`

Returns the new endpoint.  The content posted and content-type will be returned by endpoint when given the identifier and the same query string parameters.

### Adds to or updates endpoint
`POST /mock/{endpoint}`

Returns the given endpoint.  The content posted and content-type will be returned by endpoint when given the identifier and the same query string parameters.

### Make request to endpoint
`GET /endpoint/{endpoint}`

Returns the content-type and content for the endpoint with the same query string parameters

Returns 404 if endpoint does not exist

Returns 400 if endpoint exists but no content with the given query string parameters exist

### Simple response code
`GET /response/{status:int}`

Returns the status code specified
