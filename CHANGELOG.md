# Change Log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]

## Added

- CORS headers
- Firebase JWT Validation
- Restrict access by user

## Changed

- Update to Go 1.8
- Command line configuration instead of environment variables

## Fixed

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

[Unreleased]:https://github.com/LewisWatson/carshare-back/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/LewisWatson/carshare-back/releases/tag/v0.1.0
[0.2.0]: https://github.com/LewisWatson/carshare-back/releases/tag/v0.2.0
