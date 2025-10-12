# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [3.2.0] - 2025-01-08

### Added
- `DEFAULT_LANGUAGE` environment variable for configurable default bot language
- Support for setting default language to `en` (English) or `ru` (Russian)
- `build-release.sh` script for multi-platform Docker image building
- `purchase_test.go` test file for database purchase operations

### Fixed
- Dockerfile ARG duplication - replaced second `TARGETOS` with correct `TARGETARCH`
- Docker Compose restart policy improved from `always` to `unless-stopped`

### Changed
- Translation manager now accepts default language parameter during initialization
- Config initialization includes default language from environment variable

### Documentation
- Updated README.md with `DEFAULT_LANGUAGE` environment variable description
- Added usage examples for language configuration

## [3.1.4] - Previous Release

### Fixed
- Tribute payment processing issues

## [3.1.3] - Previous Release

### Fixed
- CryptoPay bot error in payment request handling

---

## Release Types

- **Added** for new features
- **Changed** for changes in existing functionality
- **Deprecated** for soon-to-be removed features
- **Removed** for now removed features
- **Fixed** for any bug fixes
- **Security** for vulnerability fixes

## Versioning

This project follows [Semantic Versioning](https://semver.org/):
- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes
