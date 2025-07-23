# `/cmd` Directory

This directory contains the main applications for this project. Each application has its own directory.

## Structure

* `/serviceability` - Main application entry point for the serviceability service
* **Note**: Migration tools have been removed for safety - use database admin tools for schema changes

## Guidelines

* Each application should have a small `main.go` file that only initializes and starts the application
* Business logic should be placed in the `/internal` packages
* Keep dependencies minimal in the main applications
* Each application should handle its own configuration using the shared config package

## Usage

To run the main service:
```
go run cmd/serviceability/main.go
```

For database schema changes, use database administration tools directly rather than automated migrations for safety. 