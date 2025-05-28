# `/cmd` Directory

This directory contains the main applications for this project. Each application has its own directory.

## Structure

* `/migrations` - Database migration tool for creating and seeding the database schema
* `/serviceability` - Main application entry point for the serviceability service

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

To run migrations:
```
go run cmd/migrations/main.go
``` 