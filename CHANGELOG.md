# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [0.3.0] - 2025-05-18

### Added
* The `clone` command does support recursive structures for groups using the `style` argument. An environment-variable `REPOW_STYLE` can also be set, to set the default behaviour.
* Providing an experimental windows binary.
* Support for self-hosted Gitlab installations. The Gitlab-Host can be changed via the environment Variable `REPOW_GITLAB_HOST`.


## [0.2.3] - 2022-11-09

### Added
* Added binaries for Mac arm64 architecture

### Changed
* Unified binary filenames

### Fixed
* Fix error handling in cleanup task



## [0.2.2] - 2022-05-27

### Changed
* Decreased default parallelism from 64 to 32

### Fixed
* Repository detection for names with dot in `cleanup`



## [0.2.1] - 2021-08-18

### Added
* Option for own slack prefix/emoji in notifications via `REPOW_SLACK_PREFIX`
* Option for retries when downloading the repo.yaml file `REPOW_GITLAB_DOWNLOAD_RETRIES`/`GITLAB_DOWNLOAD_RETRIES`



## [0.2.0] - 2021-06-20

### Added
* Also support `REPOW_GITLAB_SECRET_TOKEN` env variable additional to `GITLAB_SECRET_TOKEN`
* Quite option for validate command
* Support for `gitlab.forking_access_level`
* Optional contacts are allowed with the option `-c` in validate/apply and via `REPOW_OPTIONAL_CONTACTS` env variable in serve
* Integrated slack for serve command, to get notifications on invalid manifest files.

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



[Unreleased]: https://github.com/galan/repow/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/galan/repow/compare/v0.2.3...v0.3.0
[0.2.3]: https://github.com/galan/repow/compare/v0.2.2...v0.2.3
[0.2.2]: https://github.com/galan/repow/compare/v0.2.1...v0.2.2
[0.2.1]: https://github.com/galan/repow/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/galan/repow/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/galan/repow/compare/v0.0.2...v0.1.0
[0.0.2]: https://github.com/galan/repow/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/galan/repow/releases/tag/v0.0.1
