Mocker
======

A simple service to mock rest apis

Mocker respects and enforces query string parameters.  Two requests to the same endpoint with different query string parameters may return different content as long as content is set for both sets of parameters.

Installation
------------

~~~~
# Choose project directory or skip if GOPATH already setup
export GOPATH=~/Documents/go
mkdir -p $GOPATH

go get -v github.com/dmcgowan/mocker
$GOPATH/bin/mocker -port 8080
~~~~


API
---

### Create a new endpoint
`POST /mock[?optional query parameters]`

Returns a new endpoint identifier.  The content posted and content-type will be returned by endpoint when given the identifier, no additional path, and using the same query parameters.  Additional content may be added to the endpoint by posting updates with different path/query parameters.

### Adds to or updates endpoint
`POST /mock/{endpoint}[/optional path][?optional query parameters)]`

Returns the given endpoint.  The content posted and content-type will be returned by endpoint when given the identifier and the same path/query parameters.

### Make request to endpoint
`GET /endpoint/{endpoint}[/optional path][?optional query parameters)]`

Returns the content-type and content for the endpoint with the same path/query parameters

Returns 404 if endpoint does not exist

Returns 400 if endpoint exists but no content with the given path/query parameters exist

### Endpoint Settings

`POST /settings/{endpoint}`

Values (Form or Query)
 - `latency`
   - static: Use `latency_ms` for this endpoint
   - normal: Use normal random latency distribution using `latency_min_ms` `latency_median_ms` and `latency_max_ms`
   - none: No latency
 - `latency_ms` static latency value in milliseconds
 - `latency_min_ms`  Lower bound (approx 1%) of normal latency in milliseconds
 - `latency_median_ms` Median value of normal latency in milliseconds
 - `latency-max_ms` Upper bound (approx 1%)  of normal latency in milliseconds


### Simple response code
`GET /response/{status:int}[/optional path]`

Returns the status code specified

### Simple timeout/latency
`GET /timeout/{timeout:int}[/optional path]`

Returns 200 after specified timeout in ms

### Normal distribution timeout/latency

`GET /timeout/{min:int},{median:int},{max:int}[/optional path]`

Returns 200 after a random amount of time on a normal distribution
 - approximately 1% of requests will be min
 - approximately 1% of requests will be max
 - approximately 75% of requests will be closer to median than min or max

### Normal Latency Distribution
Example histogram of a normal distribution for 500 random latencies with minimum 4ms, median 20ms, and max 80ms
~~~~
|
|                   *
|                   *
|                   *
|              * ** *
|              * ** *
|              * ****
|              * ****
|              * ****
|             ** ****
|             *******
|             *******
|             *******
|             *******
|           * *******
|           * *******
|           * ********
|        * ** ********
|        * ***********     *        *
|        * ************    *        *
|        * ************    *        *      *          *
|        * ************    *        *      *          *
|        * ************    **       *      *          *
|        * ************    *** * *  *   *  * * *      *                          *
|    *   ****************  *** * ** *   **** * *  * * *                          *
|    ** *****************  ***** ****   **** * ** * * *                          *
|    ** ******************************  **** **** * ***    *                     *
|    ** ***************************************** ***** *  * *                   *
|    **********************************************************           * *    *
|    ********************************************************** *      ** ***   **
|_________________________________________________________________________________
     |               |                                                           |
    4ms             20ms                                                        80ms
    Min             Median                                                      Max
~~~~

