package di

import (
	"prayog-serviceability-service/internal/httpapi"
	"prayog-serviceability-service/internal/infrastructure/dbconn"
	"prayog-serviceability-service/internal/infrastructure/logger"
	"prayog-serviceability-service/internal/repository"
	"prayog-serviceability-service/internal/service/usecase"
)

// Container represents a dependency injection container.
type Container struct {
	// Logger
	Logger logger.Logger

	// Database
	DB dbconn.Database

	// Repositories
	Repositories *repository.Repositories

	// Services
	Services *usecase.Services

	// HTTP Server
	HTTPServer *httpapi.Server
}

// NewContainer creates a new DI container.
func NewContainer() (*Container, error) {
	container := &Container{}

	// Initialize components using providers
	if err := container.provideLogger(); err != nil {
		return nil, err
	}

	if err := container.provideDatabase(); err != nil {
		return nil, err
	}

	if err := container.provideRepositories(); err != nil {
		return nil, err
	}

	if err := container.provideServices(); err != nil {
		return nil, err
	}

	if err := container.provideHTTPServer(); err != nil {
		return nil, err
	}

	return container, nil
}

// Cleanup releases any resources held by the container.
func (c *Container) Cleanup() error {
	// Add cleanup logic here
	return nil
}
