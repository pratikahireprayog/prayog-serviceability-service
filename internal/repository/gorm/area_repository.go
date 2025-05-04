package gorm

import (
	"context"

	"prayog-serviceability-service/internal/domain"
	"prayog-serviceability-service/pkg/database"
)

// AreaRepository implements domain.AreaRepository using GORM
type AreaRepository struct {
	*Repository
}

// NewAreaRepository creates a new area repository
func NewAreaRepository(db *database.DB) domain.AreaRepository {
	return &AreaRepository{
		Repository: NewRepository(db),
	}
}

// GetByID retrieves an area by its ID
func (r *AreaRepository) GetByID(ctx context.Context, id uint) (*domain.Area, error) {
	var area database.Area
	if err := r.WithContext(ctx).First(&area, id).Error; err != nil {
		return nil, HandleError(err)
	}
	return mapDatabaseAreaToDomain(&area), nil
}

// GetByCityID retrieves areas by city ID
func (r *AreaRepository) GetByCityID(ctx context.Context, cityID uint) ([]*domain.Area, error) {
	var areas []database.Area
	if err := r.WithContext(ctx).Where("city_id = ?", cityID).Find(&areas).Error; err != nil {
		return nil, HandleError(err)
	}

	domainAreas := make([]*domain.Area, len(areas))
	for i, area := range areas {
		domainAreas[i] = mapDatabaseAreaToDomain(&area)
	}
	return domainAreas, nil
}

// List retrieves all areas
func (r *AreaRepository) List(ctx context.Context) ([]*domain.Area, error) {
	var areas []database.Area
	if err := r.WithContext(ctx).Find(&areas).Error; err != nil {
		return nil, HandleError(err)
	}

	domainAreas := make([]*domain.Area, len(areas))
	for i, area := range areas {
		domainAreas[i] = mapDatabaseAreaToDomain(&area)
	}
	return domainAreas, nil
}

// Create creates a new area
func (r *AreaRepository) Create(ctx context.Context, area *domain.Area) error {
	dbArea := mapDomainAreaToDatabase(area)
	if err := r.WithContext(ctx).Create(&dbArea).Error; err != nil {
		return HandleError(err)
	}
	// Update the ID after creation
	area.ID = dbArea.ID
	return nil
}

// Update updates an existing area
func (r *AreaRepository) Update(ctx context.Context, area *domain.Area) error {
	dbArea := mapDomainAreaToDatabase(area)
	result := r.WithContext(ctx).Save(&dbArea)
	if result.Error != nil {
		return HandleError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete deletes an area by its ID
func (r *AreaRepository) Delete(ctx context.Context, id uint) error {
	result := r.WithContext(ctx).Delete(&database.Area{}, id)
	if result.Error != nil {
		return HandleError(result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// mapDatabaseAreaToDomain maps a database area to a domain area
func mapDatabaseAreaToDomain(dbArea *database.Area) *domain.Area {
	return &domain.Area{
		ID:        dbArea.ID,
		CityID:    dbArea.CityID,
		Name:      dbArea.Name,
		CreatedAt: dbArea.CreatedAt,
		UpdatedAt: dbArea.UpdatedAt,
	}
}

// mapDomainAreaToDatabase maps a domain area to a database area
func mapDomainAreaToDatabase(domainArea *domain.Area) database.Area {
	return database.Area{
		ID:        domainArea.ID,
		CityID:    domainArea.CityID,
		Name:      domainArea.Name,
		CreatedAt: domainArea.CreatedAt,
		UpdatedAt: domainArea.UpdatedAt,
	}
}
