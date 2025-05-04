# `/pkg` Directory

The `/pkg` directory contains code that can be imported and used by external applications. This is the place for code that's reusable across multiple applications.

## Guidelines

* Place only reusable, public-facing code in this directory
* Keep dependencies minimal
* Document exported functions, types, and constants thoroughly
* Write comprehensive tests for all code

## Current Packages

* (None yet, will include client packages and shared libraries)

## Planned Packages

* `client` - Go client for interacting with the serviceability service
* `models` - Shared data models for the serviceability service
* `errors` - Common error types and error handling utilities 