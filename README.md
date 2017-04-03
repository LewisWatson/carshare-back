# Car Share Back
[![Build Status](https://travis-ci.org/LewisWatson/carshare-back.svg?branch=master)](https://travis-ci.org/LewisWatson/carshare-back)
[![Coverage Status](https://coveralls.io/repos/github/LewisWatson/carshare-back/badge.svg?branch=feature%2Ffirebase-jwt-auth)](https://coveralls.io/github/LewisWatson/carshare-back?branch=feature%2Ffirebase-jwt-auth)
[![GitPitch](https://gitpitch.com/assets/badge.svg)](https://gitpitch.com/LewisWatson/carshare-ninja-pitch/master?grs=github&t=white)
[![stability-experimental](https://img.shields.io/badge/stability-experimental-orange.svg)](https://github.com/emersion/stability-badges#experimental)

An Open Source API for tracking car shares.

Designed to enable easy tracking of the distance members travel as passengers and as drivers. Car share members can make informed day to day decisions about who should drive next based on the ratio of distance travelled as the driver vs as a passenger.

Written in [Go] and designed using [{json:api}] specification for building API's in JSON.

## Install

```bash
go get github.com/LewisWatson/carshare-back
```

## Run

```bash
docker run -d -p 27017:27017 mongo --smallfiles
$GOPATH/bin/carshare-back
1970/01/01 00:00:00 connecting to mongodb server via url: localhost
1970/01/01 00:00:00 listening on :31415
```

### Configuration

```bash
$GOPATH/bin/carshare-back --help
usage: carshare-back [<flags>]

API for tracking car shares

Flags:
  --help                        Show context-sensitive help (also try --help-long and --help-man).
  --port=31415                  Set port to bind to
  --mgoURL=localhost            URL to MongoDB server or seed server(s) for clusters
  --firebase="ridesharelogger"  Firebase project to use for authentication
  --cors=URI                    Enable HTTP Access Control (CORS) for the specified URI
  --version                     Show application version.
```

## Docker

The [Dockerfile](Dockerfile) uses a [minimal Docker image based on Alpine Linux](https://hub.docker.com/_/alpine/) with a different implimentation of libc. Therefore, it is important that a static binary is used when building

```bash
go build --tags netgo --ldflags '-extldflags "-lm -lstdc++ -static"'
docker build .
```

Pre-made images are available as [lewiswatson/carshare-back](https://hub.docker.com/r/lewiswatson/carshare-back/)

## License

Copyright 2017 Lewis Watson

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
[{json:api}]: http://jsonapi.org
[Go]: https://golang.org/
