# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Version management package for tracking service versioning
- `/version` endpoint to expose service version information
- API version header (`X-API-Version`) to all responses
- Makefile support for injecting version information at build time
- This CHANGELOG.md file

## [1.0.0] - YYYY-MM-DD

### Added

- Initial release of Prayog Serviceability Service
- Serviceability check endpoint to verify if a location is serviceable
- Bulk serviceability check endpoint
- Support for handling postal codes, regions, cities, and countries
- GORM-based repository pattern implementation
- GoFiber web framework integration
- Health check endpoint
- Configuration management using Viper
- Logging using zap
