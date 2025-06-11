package repositories

import (
	"context"
	"fmt"
	"strings"

	"prayog-serviceability-service/internal/shared/models/v1"

	"gorm.io/gorm"
)

// partnerLocationCoverageRepository implements PartnerLocationCoverageRepository
type partnerLocationCoverageRepository struct {
	db *gorm.DB
}

// NewPartnerLocationCoverageRepository creates a new partner location coverage repository
func NewPartnerLocationCoverageRepository(db *gorm.DB) PartnerLocationCoverageRepository {
	return &partnerLocationCoverageRepository{db: db}
}

// GetByPartnerID retrieves all coverage records for a partner
func (r *partnerLocationCoverageRepository) GetByPartnerID(ctx context.Context, partnerID uint) ([]models.PartnerLocationCoverage, error) {
	var coverages []models.PartnerLocationCoverage

	err := r.db.WithContext(ctx).
		Where("partner_id = ? AND is_active = ?", partnerID, true).
		Find(&coverages).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get partner coverage: %w", err)
	}

	return coverages, nil
}

// GetByLocationScopeAndID retrieves coverage records for a specific location
func (r *partnerLocationCoverageRepository) GetByLocationScopeAndID(ctx context.Context, locationScope string, locationID uint) ([]models.PartnerLocationCoverage, error) {
	var coverages []models.PartnerLocationCoverage

	err := r.db.WithContext(ctx).
		Where("location_scope = ? AND location_id = ? AND is_active = ?",
			strings.ToUpper(locationScope), locationID, true).
		Find(&coverages).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get location coverage: %w", err)
	}

	return coverages, nil
}

// GetByFilters retrieves coverage records with location hierarchy based on filters
func (r *partnerLocationCoverageRepository) GetByFilters(ctx context.Context, filters *models.PartnerLocationCoverageFilters) ([]models.PartnerLocationCoverageResult, error) {
	var results []models.PartnerLocationCoverageResult

	query := r.db.WithContext(ctx).Table("partner_location_coverages plc")

	// Apply filters
	if filters.PartnerID != nil {
		query = query.Where("plc.partner_id = ?", *filters.PartnerID)
	}
	if filters.LocationScope != "" {
		query = query.Where("plc.location_scope = ?", strings.ToUpper(filters.LocationScope))
	}
	if filters.LocationID != nil {
		query = query.Where("plc.location_id = ?", *filters.LocationID)
	}
	if filters.ZoneType != "" {
		query = query.Where("plc.zone_type = ?", strings.ToUpper(filters.ZoneType))
	}
	if filters.SourcePostalCode != nil {
		query = query.Where("plc.source_postal_code = ?", *filters.SourcePostalCode)
	}
	if filters.DestinationPostalCode != nil {
		query = query.Where("plc.destination_postal_code = ?", *filters.DestinationPostalCode)
	}
	if filters.SourcePostalCodeID != nil {
		query = query.Where("plc.source_postal_code_id = ?", *filters.SourcePostalCodeID)
	}
	if filters.DestinationPostalCodeID != nil {
		query = query.Where("plc.destination_postal_code_id = ?", *filters.DestinationPostalCodeID)
	}
	if filters.IsActive != nil {
		query = query.Where("plc.is_active = ?", *filters.IsActive)
	} else {
		query = query.Where("plc.is_active = ?", true)
	}

	rows, err := query.Select("plc.partner_id, plc.location_scope, plc.location_id, plc.zone_type").Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query partner coverage: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var result models.PartnerLocationCoverageResult
		if err := rows.Scan(&result.PartnerID, &result.LocationScope, &result.LocationID, &result.ZoneType); err != nil {
			return nil, fmt.Errorf("failed to scan coverage result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetPartnersForLocation retrieves all partners serving a specific location with zone priorities
func (r *partnerLocationCoverageRepository) GetPartnersForLocation(ctx context.Context, locationScope string, locationID uint, zoneTypes []string) ([]models.PartnerLocationCoverageResult, error) {
	var results []models.PartnerLocationCoverageResult

	query := r.db.WithContext(ctx).Table("partner_location_coverages plc").
		Where("plc.location_scope = ? AND plc.location_id = ? AND plc.is_active = ?",
			strings.ToUpper(locationScope), locationID, true)

	if len(zoneTypes) > 0 {
		upperZoneTypes := make([]string, len(zoneTypes))
		for i, zt := range zoneTypes {
			upperZoneTypes[i] = strings.ToUpper(zt)
		}
		query = query.Where("plc.zone_type IN ?", upperZoneTypes)
	}

	// Order by zone type priority (PRIMARY, SECONDARY, BUFFER)
	query = query.Order("CASE plc.zone_type WHEN 'PRIMARY' THEN 1 WHEN 'SECONDARY' THEN 2 WHEN 'BUFFER' THEN 3 END")

	rows, err := query.Select("plc.partner_id, plc.location_scope, plc.location_id, plc.zone_type").Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query partners for location: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var result models.PartnerLocationCoverageResult
		if err := rows.Scan(&result.PartnerID, &result.LocationScope, &result.LocationID, &result.ZoneType); err != nil {
			return nil, fmt.Errorf("failed to scan partner result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetCoverageForPartner retrieves all coverage areas for a specific partner
func (r *partnerLocationCoverageRepository) GetCoverageForPartner(ctx context.Context, partnerID uint, isActive bool) ([]models.PartnerLocationCoverageResult, error) {
	var results []models.PartnerLocationCoverageResult

	query := r.db.WithContext(ctx).Table("partner_location_coverages plc").
		Where("plc.partner_id = ? AND plc.is_active = ?", partnerID, isActive)

	rows, err := query.Select("plc.partner_id, plc.location_scope, plc.location_id, plc.zone_type").Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query partner coverage: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var result models.PartnerLocationCoverageResult
		if err := rows.Scan(&result.PartnerID, &result.LocationScope, &result.LocationID, &result.ZoneType); err != nil {
			return nil, fmt.Errorf("failed to scan coverage result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// Create creates a new partner location coverage record
func (r *partnerLocationCoverageRepository) Create(ctx context.Context, coverage *models.PartnerLocationCoverage) error {
	err := r.db.WithContext(ctx).Create(coverage).Error
	if err != nil {
		return fmt.Errorf("failed to create partner location coverage: %w", err)
	}
	return nil
}

// Update updates an existing partner location coverage record
func (r *partnerLocationCoverageRepository) Update(ctx context.Context, coverage *models.PartnerLocationCoverage) error {
	result := r.db.WithContext(ctx).Save(coverage)
	if result.Error != nil {
		return fmt.Errorf("failed to update partner location coverage: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("partner location coverage not found")
	}
	return nil
}

// Delete soft deletes a partner location coverage record by marking it as deleted
func (r *partnerLocationCoverageRepository) Delete(ctx context.Context, id uint) error {
	// Get the partner location coverage
	var coverage models.PartnerLocationCoverage
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&coverage).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("partner location coverage not found")
		}
		return fmt.Errorf("failed to find partner location coverage: %w", err)
	}

	// Use the model's soft delete method
	return coverage.SoftDelete(r.db.WithContext(ctx))
}

// BulkCreate creates multiple partner location coverage records in a transaction
func (r *partnerLocationCoverageRepository) BulkCreate(ctx context.Context, coverages []models.PartnerLocationCoverage) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.CreateInBatches(coverages, 100).Error; err != nil {
			return fmt.Errorf("failed to bulk create partner location coverages: %w", err)
		}
		return nil
	})
}

// GetByPartnerIDWithDeleted retrieves all coverage records for a partner (includes soft-deleted records)
func (r *partnerLocationCoverageRepository) GetByPartnerIDWithDeleted(ctx context.Context, partnerID uint) ([]models.PartnerLocationCoverage, error) {
	var coverages []models.PartnerLocationCoverage

	err := r.db.WithContext(ctx).Unscoped().
		Where("partner_id = ?", partnerID).
		Find(&coverages).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get partner coverage: %w", err)
	}

	return coverages, nil
}

// GetByLocationScopeAndIDWithDeleted retrieves coverage records for a specific location (includes soft-deleted records)
func (r *partnerLocationCoverageRepository) GetByLocationScopeAndIDWithDeleted(ctx context.Context, locationScope string, locationID uint) ([]models.PartnerLocationCoverage, error) {
	var coverages []models.PartnerLocationCoverage

	err := r.db.WithContext(ctx).Unscoped().
		Where("location_scope = ? AND location_id = ?",
			strings.ToUpper(locationScope), locationID).
		Find(&coverages).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get location coverage: %w", err)
	}

	return coverages, nil
}

// GetByFiltersWithDeleted retrieves coverage records with location hierarchy based on filters (includes soft-deleted records)
func (r *partnerLocationCoverageRepository) GetByFiltersWithDeleted(ctx context.Context, filters *models.PartnerLocationCoverageFilters) ([]models.PartnerLocationCoverageResult, error) {
	var results []models.PartnerLocationCoverageResult

	query := r.db.WithContext(ctx).Unscoped().Table("partner_location_coverages plc")

	// Apply filters
	if filters.PartnerID != nil {
		query = query.Where("plc.partner_id = ?", *filters.PartnerID)
	}
	if filters.LocationScope != "" {
		query = query.Where("plc.location_scope = ?", strings.ToUpper(filters.LocationScope))
	}
	if filters.LocationID != nil {
		query = query.Where("plc.location_id = ?", *filters.LocationID)
	}
	if filters.ZoneType != "" {
		query = query.Where("plc.zone_type = ?", strings.ToUpper(filters.ZoneType))
	}
	if filters.SourcePostalCode != nil {
		query = query.Where("plc.source_postal_code = ?", *filters.SourcePostalCode)
	}
	if filters.DestinationPostalCode != nil {
		query = query.Where("plc.destination_postal_code = ?", *filters.DestinationPostalCode)
	}
	if filters.SourcePostalCodeID != nil {
		query = query.Where("plc.source_postal_code_id = ?", *filters.SourcePostalCodeID)
	}
	if filters.DestinationPostalCodeID != nil {
		query = query.Where("plc.destination_postal_code_id = ?", *filters.DestinationPostalCodeID)
	}
	if filters.IsActive != nil {
		query = query.Where("plc.is_active = ?", *filters.IsActive)
	}

	rows, err := query.Select("plc.partner_id, plc.location_scope, plc.location_id, plc.zone_type").Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query partner coverage: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var result models.PartnerLocationCoverageResult
		if err := rows.Scan(&result.PartnerID, &result.LocationScope, &result.LocationID, &result.ZoneType); err != nil {
			return nil, fmt.Errorf("failed to scan coverage result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetPartnersForLocationWithDeleted retrieves all partners serving a specific location with zone priorities (includes soft-deleted records)
func (r *partnerLocationCoverageRepository) GetPartnersForLocationWithDeleted(ctx context.Context, locationScope string, locationID uint, zoneTypes []string) ([]models.PartnerLocationCoverageResult, error) {
	var results []models.PartnerLocationCoverageResult

	query := r.db.WithContext(ctx).Unscoped().Table("partner_location_coverages plc").
		Where("plc.location_scope = ? AND plc.location_id = ?",
			strings.ToUpper(locationScope), locationID)

	if len(zoneTypes) > 0 {
		upperZoneTypes := make([]string, len(zoneTypes))
		for i, zt := range zoneTypes {
			upperZoneTypes[i] = strings.ToUpper(zt)
		}
		query = query.Where("plc.zone_type IN ?", upperZoneTypes)
	}

	// Order by zone type priority (PRIMARY, SECONDARY, BUFFER)
	query = query.Order("CASE plc.zone_type WHEN 'PRIMARY' THEN 1 WHEN 'SECONDARY' THEN 2 WHEN 'BUFFER' THEN 3 END")

	rows, err := query.Select("plc.partner_id, plc.location_scope, plc.location_id, plc.zone_type").Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query partners for location: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var result models.PartnerLocationCoverageResult
		if err := rows.Scan(&result.PartnerID, &result.LocationScope, &result.LocationID, &result.ZoneType); err != nil {
			return nil, fmt.Errorf("failed to scan partner result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetCoverageForPartnerWithDeleted retrieves all coverage areas for a specific partner (includes soft-deleted records)
func (r *partnerLocationCoverageRepository) GetCoverageForPartnerWithDeleted(ctx context.Context, partnerID uint, isActive bool) ([]models.PartnerLocationCoverageResult, error) {
	var results []models.PartnerLocationCoverageResult

	query := r.db.WithContext(ctx).Unscoped().Table("partner_location_coverages plc").
		Where("plc.partner_id = ?", partnerID)

	// Only apply is_active filter if specified
	if isActive {
		query = query.Where("plc.is_active = ?", isActive)
	}

	rows, err := query.Select("plc.partner_id, plc.location_scope, plc.location_id, plc.zone_type").Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query partner coverage: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var result models.PartnerLocationCoverageResult
		if err := rows.Scan(&result.PartnerID, &result.LocationScope, &result.LocationID, &result.ZoneType); err != nil {
			return nil, fmt.Errorf("failed to scan coverage result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetOnlyDeleted retrieves only soft-deleted partner location coverage records with pagination
func (r *partnerLocationCoverageRepository) GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.PartnerLocationCoverage, int64, error) {
	var coverages []models.PartnerLocationCoverage
	var total int64

	// Count only soft-deleted records
	err := r.db.WithContext(ctx).Unscoped().Model(&models.PartnerLocationCoverage{}).Where("is_deleted = ?", true).Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deleted partner location coverages: %w", err)
	}

	// Get paginated soft-deleted records
	err = r.db.WithContext(ctx).Unscoped().
		Where("is_deleted = ?", true).
		Offset(offset).Limit(limit).
		Find(&coverages).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get deleted partner location coverages: %w", err)
	}

	return coverages, total, nil
}

// Restore restores a soft-deleted partner location coverage record
func (r *partnerLocationCoverageRepository) Restore(ctx context.Context, id uint) error {
	// Get the soft-deleted coverage
	var coverage models.PartnerLocationCoverage
	err := r.db.WithContext(ctx).Unscoped().Where("id = ? AND is_deleted = ?", id, true).First(&coverage).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("soft-deleted partner location coverage not found")
		}
		return fmt.Errorf("failed to find soft-deleted partner location coverage: %w", err)
	}

	// Use the model's restore method
	return coverage.Restore(r.db.WithContext(ctx))
}

// ForceDelete permanently deletes a partner location coverage record (hard delete)
func (r *partnerLocationCoverageRepository) ForceDelete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Unscoped().Delete(&models.PartnerLocationCoverage{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to force delete partner location coverage: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("partner location coverage not found")
	}
	return nil
}
