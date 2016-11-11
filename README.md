# Car Share Back
[![Build Status](https://travis-ci.org/LewisWatson/carshare-back.svg?branch=master)](https://travis-ci.org/LewisWatson/carshare-back)
[![stability-experimental](https://img.shields.io/badge/stability-experimental-orange.svg)](https://github.com/emersion/stability-badges#experimental)

An Open Source API for tracking car shares.

Designed to enable easy tracking of the distance members travel as passengers and as drivers. Car share members can make informed day to day decisions about who should drive next based on the ratio of distance travelled as driver vs as passenger.

Written in Go and confirming to [json:api specification](http://jsonapi.org) for a consistent [HATEOAS](https://en.wikipedia.org/wiki/HATEOAS) REST interface.

## Install

```bash
$ go get github.com/LewisWatson/carshare-back
```

## Run

```bash
$ $GOPATH/bin/carshare-back
Listening on :31415
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
