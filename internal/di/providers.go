package di

import (
	"prayog-serviceability-service/internal/business"
	"prayog-serviceability-service/internal/httpapi"
	"prayog-serviceability-service/internal/infrastructure/dbconn"
	"prayog-serviceability-service/internal/infrastructure/logger"
	"prayog-serviceability-service/internal/repository"
)

// provideLogger initializes and provides the logger.
func (c *Container) provideLogger() error {
	// Initialize logger
	log, err := logger.NewLogger()
	if err != nil {
		return err
	}
	c.Logger = log
	return nil
}

// provideDatabase initializes and provides the database.
func (c *Container) provideDatabase() error {
	// Initialize database
	database, err := dbconn.NewDatabase()
	if err != nil {
		return err
	}
	c.DB = database
	return nil
}

// provideRepositories initializes and provides all repositories.
func (c *Container) provideRepositories() error {
	// Initialize repositories
	repos, err := repository.NewRepositories(c.DB)
	if err != nil {
		return err
	}
	c.Repositories = repos
	return nil
}

// provideServices initializes and provides all services.
func (c *Container) provideServices() error {
	// Initialize services
	services, err := business.NewServices(c.Repositories, c.Logger)
	if err != nil {
		return err
	}
	c.Services = services
	return nil
}

// provideHTTPServer initializes and provides the HTTP server.
func (c *Container) provideHTTPServer() error {
	// Initialize HTTP server
	server, err := httpapi.NewServer(c.Services, c.Logger)
	if err != nil {
		return err
	}
	c.HTTPServer = server
	return nil
}
