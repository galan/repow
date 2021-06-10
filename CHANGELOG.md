# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [Unreleased]

### Fixed
* Install CA-certs (for webhook requests)



## [0.1.0] - 2021-06-06

### Added
* Improved parallelism by limiting go-routines (thanks to [@flecno](https://github.com/flecno))
* Added Tests for matching
* Added CHANGELOG.md

### Changed
* Configurable parallelism by flag
* Unified duration log

### Fixed
* Adjusted default server port



## [0.0.2] - 2021-06-03

### Added
* Cleanup docs

### Fixed
* Changed behaviour when files in git parent root



## [0.0.1] - 2021-06-03

### Added
* First public release



[Unreleased]: https://github.com/galan/repow/compare/v0.1.0...HEAD
[0.0.2]: https://github.com/galan/repow/compare/v0.0.2...v0.1.0
[0.0.2]: https://github.com/galan/repow/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/galan/repow/releases/tag/v0.0.1
