# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

### Changed

- Introduce dependency management with golang/dep after run in with dependecncy hell

## [0.4.1] - 2017-05-03

### Fixed

- Ensure the database connection is always stored in the context, not just
  when CORS is enabled.

## [0.4.0] - 2017-04-04

### Added

- Prometheus metrics collection via `/metrics` endpoint

### Changed

- Add access control to trip resource
- New trips need to have a valid car share on creation
- Switch to Gin Framework
- Switch to https://github.com/op/go-logging for more flexible logging
- Make firebase project configurable

## [0.3.3] - 2017-03-27

### Fixed

- Use full mgoURL from command line to connect to database

## [0.3.2] - 2017-03-23

### Changed

- Switch base docker image to alpine, reduces image size from 200mb+ to less than 20mb
- Command line arguments working via docker run
  ```bash
  docker run carshare-back --port 1337 --mgoURL mongo
  2017/03/23 07:25:55 connecting to mongodb server via url: mongo
  2017/03/23 07:25:56 listening on :1337
  ```

## [0.3.1] - 2017-03-21

### Fixed

- Re-introduce environmental variable configuration support (for docker)

## [0.3.0] - 2017-03-21

### Added

- CORS headers
- Firebase JWT Validation
- Restrict access by user

### Changed

- Update to Go 1.8
- Command line configuration instead of environment variables

### Fixed

- Relation links between car shares, trips and users

## [0.2.0] - 2016-12-10

### Added

- MongoDB data store support
- Configure MongoDB URL via `CARSHARE_MGO_URL` environment variable
- Configure server port via `CARSHARE_PORT` environment variable
- Created Dockerfile

### Changed

- Standardisd ID's on [BSON ObjectId](https://docs.mongodb.com/manual/reference/bson-types/#objectid)
- Overhaul error handling
- Unit test now run twice. First with the fast in-memory data store (fail fast), then as an integration test against a MongoDB docker container


## [0.1.0] - 2016-11-12

### Added

- Create basic functionality with in memory data store
- Add ability to create users, car shares and trips via json:api REST interface and store in simple in memory data store
- Add README and CHANGELOG

[Unreleased]:https://github.com/LewisWatson/carshare-back/compare/v0.4.1...HEAD
[0.4.1]: https://github.com/LewisWatson/carshare-back/releases/tag/v0.4.1
[0.4.0]: https://github.com/LewisWatson/carshare-back/releases/tag/v0.4.0
[0.3.3]: https://github.com/LewisWatson/carshare-back/releases/tag/v0.3.3
[0.3.2]: https://github.com/LewisWatson/carshare-back/releases/tag/v0.3.2
[0.3.1]: https://github.com/LewisWatson/carshare-back/releases/tag/v0.3.1
[0.3.0]: https://github.com/LewisWatson/carshare-back/releases/tag/v0.3.0
[0.2.0]: https://github.com/LewisWatson/carshare-back/releases/tag/v0.2.0
[0.1.0]: https://github.com/LewisWatson/carshare-back/releases/tag/v0.1.0
