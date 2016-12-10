# Car Share Back
[![Build Status](https://travis-ci.org/LewisWatson/carshare-back.svg?branch=master)](https://travis-ci.org/LewisWatson/carshare-back)
[![stability-experimental](https://img.shields.io/badge/stability-experimental-orange.svg)](https://github.com/emersion/stability-badges#experimental)

An Open Source API for tracking car shares.

Designed to enable easy tracking of the distance members travel as passengers and as drivers. Car share members can make informed day to day decisions about who should drive next based on the ratio of distance travelled as the driver vs as a passenger.

Written in Go and designed using [{json:api}] specification for building API's in JSON.

## Install

```bash
go get github.com/LewisWatson/carshare-back
```

## Run

```bash
$GOPATH/bin/carshare-back
1970/01/01 00:00:00 connecting to mongodb server via url: localhost
1970/01/01 00:00:00 listening on :31415
```

### Configuration

#### [MongoDB]([mongoDB]) Data Store
Carshare-back uses [mongoDB] as a data store. By default it will look for one running on `localhost`. You set an [alternative url](https://godoc.org/labix.org/v2/mgo#Dial) via the `CARSHARE_MGO_URL` environment variable.

#### Port
The default port is `31415`. You can set an alternative via the `CARSHARE_PORT` environment variable.

## Docker
A [Dockerfile](Dockerfile) is provided for generating a `carshare-back` docker container.

###  Build

```bash
docker build -t carshare-back .
```

### Run

First, run a [mongoDB] container

```
docker run -d --name mongo mongo
```

Then, run `carshare-back` with a link to the mongoDB container.

```bash
docker run --link mongo:mongo -p 31415:31415 carshare-back
```

## License

Copyright 2016 Lewis Watson

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

[mongoDB]: https://www.mongodb.com/
[{json:api}]: (http://jsonapi.org)
