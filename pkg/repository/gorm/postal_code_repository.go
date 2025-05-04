package gorm

import (
	"context"

	"prayog-serviceability-service/pkg/domain"
	database "prayog-serviceability-service/pkg/infrastructure/db"
	"prayog-serviceability-service/pkg/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PostalCodeRepository implements repository.PostalCodeRepository using GORM
type PostalCodeRepository struct {
	*repository.BaseRepository[domain.PostalCode, database.PostalCode]
}

// NewPostalCodeRepository creates a new postal code repository
func NewPostalCodeRepository(db *database.DB) *PostalCodeRepository {
	return &PostalCodeRepository{
		BaseRepository: repository.NewBaseRepository[domain.PostalCode, database.PostalCode](db),
	}
}

// GetByID retrieves a postal code by its ID
func (r *PostalCodeRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.PostalCode, error) {
	return r.BaseRepository.GetByID(ctx, id, mapDatabasePostalCodeToDomain)
}

// GetByCode retrieves a postal code by its code
func (r *PostalCodeRepository) GetByCode(ctx context.Context, code string) (domain.PostalCode, error) {
	return r.BaseRepository.FindOne(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Where("code = ?", code)
	}, mapDatabasePostalCodeToDomain)
}

// GetByAreaID retrieves all postal codes by area ID
func (r *PostalCodeRepository) GetByAreaID(ctx context.Context, areaID uuid.UUID) ([]domain.PostalCode, error) {
	return r.BaseRepository.Query(ctx, func(db *gorm.DB) *gorm.DB {
		return db.Where("area_id = ?", areaID)
	}, mapDatabasePostalCodeToDomain)
}

// List retrieves all postal codes
func (r *PostalCodeRepository) List(ctx context.Context) ([]domain.PostalCode, error) {
	return r.BaseRepository.List(ctx, mapDatabasePostalCodeToDomain)
}

// Create creates a new postal code
func (r *PostalCodeRepository) Create(ctx context.Context, postalCode domain.PostalCode) error {
	return r.BaseRepository.Create(ctx, postalCode, mapDomainPostalCodeToDatabase)
}

// Update updates an existing postal code
func (r *PostalCodeRepository) Update(ctx context.Context, postalCode domain.PostalCode) error {
	return r.BaseRepository.Update(ctx, postalCode, mapDomainPostalCodeToDatabase)
}

// Delete deletes a postal code by its ID
func (r *PostalCodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.BaseRepository.Delete(ctx, id, database.PostalCode{})
}

// Helper functions to map between domain and database models

func mapDatabasePostalCodeToDomain(dbPostalCode *database.PostalCode) domain.PostalCode {
	return domain.PostalCode{
		ID:          dbPostalCode.ID,
		Code:        dbPostalCode.Code,
		CountryID:   uuid.UUID{}, // Simplified mapping, should include proper fields
		RegionID:    uuid.UUID{}, // Simplified mapping
		CityID:      uuid.UUID{}, // Simplified mapping
		AreaID:      dbPostalCode.AreaID,
		GeoLocation: domain.Point{}, // Simplified mapping
		IsActive:    true,           // Default value, should be stored in DB
		CreatedAt:   dbPostalCode.CreatedAt,
		UpdatedAt:   dbPostalCode.UpdatedAt,
	}
}

func mapDomainPostalCodeToDatabase(domainPostalCode domain.PostalCode) database.PostalCode {
	return database.PostalCode{
		ID:        domainPostalCode.ID,
		Code:      domainPostalCode.Code,
		AreaID:    domainPostalCode.AreaID,
		CreatedAt: domainPostalCode.CreatedAt,
		UpdatedAt: domainPostalCode.UpdatedAt,
	}
}
