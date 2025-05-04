package repository

import (
	"prayog-serviceability-service/internal/infrastructure/dbconn"
)

// Repositories contains all data repositories.
type Repositories struct {
	// Add individual repositories as needed
	// Example: UserRepository, ProductRepository, etc.
}

// NewRepositories initializes and returns all repositories.
func NewRepositories(db dbconn.Database) (*Repositories, error) {
	return &Repositories{
		// Initialize all repositories here
	}, nil
}
