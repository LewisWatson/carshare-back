# Car Share Backend
RESTful API for Car Share System.

Written in Go using the [github.com/manyminds/api2go](https://github.com/manyminds/api2go) library, which is a [JSON API](http://jsonapi.org) Implementation for Go.

## Install

```go
go get github.com/LewisWatson/carshare-back
```

## Run

```bash
$ $GOPATH/bin/carshare-back
Listening on :31415
```


## cURL Commands

### Trips

#### Create a new trip

```bash
curl -X POST http://localhost:31415/v0/trips -d '{"data" : {"type" : "trips" , "attributes": {"metres-as-driver" : 0, "metres-as-passenger" : 0 }}}'
```

#### List trips
```bash
curl -X GET http://localhost:31415/v0/trips
```

#### List paginated trips
```bash
curl -X GET 'http://localhost:31415/v0/trips?page\[offset\]=0&page\[limit\]=2'
```
OR
```bash
curl -X GET 'http://localhost:31415/v0/trips?page\[number\]=1&page\[size\]=2'
```

#### Update a trip
```bash
curl -vX PATCH http://localhost:31415/v0/trips/1 -d '{ "data" : {"type" : "trips", "id": "1", "attributes": {"metres-as-driver" : 1, "metres-as-passenger" : 2}}}'
```

#### Delete a trip
```bash
curl -vX DELETE http://localhost:31415/v0/trips/2
```

### Car Shares

#### Create a new Car Share
```bash
	curl -X POST http://localhost:31415/v0/carShares -d '{"data" : {"type" : "carShares" , "attributes": {"name" : "Car Share 1", "metres" : 2000 }}}'
```

#### List Car Shares
```bash
curl -X GET http://localhost:31415/v0/carShares
```

#### Update a Car Share
```bash
curl -vX PATCH http://localhost:31415/v0/carShares/1 -d '{ "data" : {"type" : "carShares", "id": "1", "attributes": {"metres-as-driver" : 1, "metres-as-passenger" : 2}}}'
```

#### Delete a Car Share
```bash
curl -vX DELETE http://localhost:31415/v0/carShares/1
```
