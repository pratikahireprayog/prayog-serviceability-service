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
func (r *partnerLocationCoverageRepository) GetByPartnerID(ctx context.Context, partnerID string) ([]models.PartnerLocationCoverage, error) {
	var coverages []models.PartnerLocationCoverage

	err := r.db.WithContext(ctx).
		Where("partner_id = ? AND is_active = ?", partnerID, true).
		Find(&coverages).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get partner coverage: %w", err)
	}

	return coverages, nil
}

// GetByPostalCode retrieves coverage records for a specific postal code
func (r *partnerLocationCoverageRepository) GetByPostalCode(ctx context.Context, postalCode string) ([]models.PartnerLocationCoverage, error) {
	var coverages []models.PartnerLocationCoverage

	err := r.db.WithContext(ctx).
		Where("postal_code = ? AND is_active = ?", postalCode, true).
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

	// Apply basic filters
	if filters.PartnerID != nil {
		query = query.Where("plc.partner_id = ?", *filters.PartnerID)
	}
	if filters.PartnerCode != nil {
		query = query.Where("plc.partner_code = ?", *filters.PartnerCode)
	}
	if filters.PostalCode != nil {
		query = query.Where("plc.postal_code = ?", *filters.PostalCode)
	}
	if filters.PostalCodeID != nil {
		query = query.Where("plc.postal_code_id = ?", *filters.PostalCodeID)
	}
	if filters.ZoneType != "" {
		query = query.Where("plc.zone_type = ?", strings.ToUpper(filters.ZoneType))
	}

	// Apply serviceability filters
	if filters.CountryCode != nil {
		query = query.Where("plc.country_code = ?", strings.ToUpper(*filters.CountryCode))
	}
	if filters.ProductType != nil {
		query = query.Where("plc.product_type = ?", *filters.ProductType)
	}
	if filters.ParcelCategory != nil {
		query = query.Where("plc.parcel_category = ?", *filters.ParcelCategory)
	}
	if filters.ServiceType != nil {
		query = query.Where("plc.service_type = ?", *filters.ServiceType)
	}
	if filters.TATDays != nil {
		query = query.Where("plc.tat_days = ?", *filters.TATDays)
	}
	if filters.Pickup != nil {
		query = query.Where("plc.pickup = ?", *filters.Pickup)
	}
	if filters.Delivery != nil {
		query = query.Where("plc.delivery = ?", *filters.Delivery)
	}
	if filters.DeliveryMode != nil {
		query = query.Where("plc.delivery_mode = ?", strings.ToLower(*filters.DeliveryMode))
	}
	if filters.CODAvailable != nil {
		query = query.Where("plc.cod_available = ?", *filters.CODAvailable)
	}
	if filters.Insurance != nil {
		query = query.Where("plc.insurance = ?", *filters.Insurance)
	}
	if filters.MinWeightKG != nil {
		query = query.Where("plc.min_weight_kg >= ?", *filters.MinWeightKG)
	}
	if filters.MaxWeightKG != nil {
		query = query.Where("plc.max_weight_kg <= ?", *filters.MaxWeightKG)
	}

	if filters.IsActive != nil {
		query = query.Where("plc.is_active = ?", *filters.IsActive)
	} else {
		query = query.Where("plc.is_active = ?", true)
	}

	rows, err := query.Select("plc.partner_id, plc.postal_code, plc.postal_code_id, plc.zone_type").Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query partner coverage: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var result models.PartnerLocationCoverageResult
		if err := rows.Scan(&result.PartnerID, &result.PostalCode, &result.PostalCodeID, &result.ZoneType); err != nil {
			return nil, fmt.Errorf("failed to scan coverage result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetPartnersForLocation retrieves all partners serving a specific location with zone priorities (interface compatibility)
func (r *partnerLocationCoverageRepository) GetPartnersForLocation(ctx context.Context, locationScope string, locationID uint, zoneTypes []string) ([]models.PartnerLocationCoverageResult, error) {
	// For backward compatibility, if locationScope is "POSTAL_CODE", treat locationID as postal code
	if strings.ToUpper(locationScope) == "POSTAL_CODE" {
		return r.GetPartnersForPostalCode(ctx, fmt.Sprintf("%d", locationID), locationID, zoneTypes)
	}

	// For other location scopes, return empty result since the new schema only supports postal codes
	return []models.PartnerLocationCoverageResult{}, nil
}

// GetCoverageForPartner retrieves all coverage areas for a specific partner
func (r *partnerLocationCoverageRepository) GetCoverageForPartner(ctx context.Context, partnerID string, isActive bool) ([]models.PartnerLocationCoverageResult, error) {
	var results []models.PartnerLocationCoverageResult

	query := r.db.WithContext(ctx).Table("partner_location_coverages plc").
		Where("plc.partner_id = ? AND plc.is_active = ?", partnerID, isActive)

	rows, err := query.Select("plc.partner_id, plc.postal_code, plc.postal_code_id, plc.zone_type").Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query partner coverage: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var result models.PartnerLocationCoverageResult
		if err := rows.Scan(&result.PartnerID, &result.PostalCode, &result.PostalCodeID, &result.ZoneType); err != nil {
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
func (r *partnerLocationCoverageRepository) Delete(ctx context.Context, id string) error {
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

// Restore restores a soft-deleted partner location coverage record
func (r *partnerLocationCoverageRepository) Restore(ctx context.Context, id string) error {
	var coverage models.PartnerLocationCoverage
	err := r.db.WithContext(ctx).Unscoped().Where("id = ?", id).First(&coverage).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("partner location coverage not found")
		}
		return fmt.Errorf("failed to find partner location coverage: %w", err)
	}

	return coverage.Restore(r.db.WithContext(ctx))
}

// ForceDelete permanently deletes a partner location coverage record
func (r *partnerLocationCoverageRepository) ForceDelete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Unscoped().Delete(&models.PartnerLocationCoverage{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to force delete partner location coverage: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("partner location coverage not found")
	}
	return nil
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
func (r *partnerLocationCoverageRepository) GetByPartnerIDWithDeleted(ctx context.Context, partnerID string) ([]models.PartnerLocationCoverage, error) {
	var coverages []models.PartnerLocationCoverage

	err := r.db.WithContext(ctx).Unscoped().
		Where("partner_id = ?", partnerID).
		Find(&coverages).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get partner coverage: %w", err)
	}

	return coverages, nil
}

// GetByPostalCodeWithDeleted retrieves coverage records for a specific postal code (includes soft-deleted records)
func (r *partnerLocationCoverageRepository) GetByPostalCodeWithDeleted(ctx context.Context, postalCode string) ([]models.PartnerLocationCoverage, error) {
	var coverages []models.PartnerLocationCoverage

	err := r.db.WithContext(ctx).Unscoped().
		Where("postal_code = ?", postalCode).
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
	if filters.PostalCode != nil {
		query = query.Where("plc.postal_code = ?", *filters.PostalCode)
	}
	if filters.PostalCodeID != nil {
		query = query.Where("plc.postal_code_id = ?", *filters.PostalCodeID)
	}
	if filters.ZoneType != "" {
		query = query.Where("plc.zone_type = ?", strings.ToUpper(filters.ZoneType))
	}
	if filters.IsActive != nil {
		query = query.Where("plc.is_active = ?", *filters.IsActive)
	}

	rows, err := query.Select("plc.partner_id, plc.postal_code, plc.postal_code_id, plc.zone_type").Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query partner coverage: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var result models.PartnerLocationCoverageResult
		if err := rows.Scan(&result.PartnerID, &result.PostalCode, &result.PostalCodeID, &result.ZoneType); err != nil {
			return nil, fmt.Errorf("failed to scan coverage result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetPartnersForLocationWithDeleted retrieves all partners serving a specific location with zone priorities (includes soft-deleted records) (interface compatibility)
func (r *partnerLocationCoverageRepository) GetPartnersForLocationWithDeleted(ctx context.Context, locationScope string, locationID uint, zoneTypes []string) ([]models.PartnerLocationCoverageResult, error) {
	// For backward compatibility, if locationScope is "POSTAL_CODE", treat locationID as postal code
	if strings.ToUpper(locationScope) == "POSTAL_CODE" {
		return r.GetPartnersForPostalCodeWithDeleted(ctx, fmt.Sprintf("%d", locationID), locationID, zoneTypes)
	}

	// For other location scopes, return empty result since the new schema only supports postal codes
	return []models.PartnerLocationCoverageResult{}, nil
}

// GetCoverageForPartnerWithDeleted retrieves all coverage areas for a specific partner (includes soft-deleted records)
func (r *partnerLocationCoverageRepository) GetCoverageForPartnerWithDeleted(ctx context.Context, partnerID string, isActive bool) ([]models.PartnerLocationCoverageResult, error) {
	var results []models.PartnerLocationCoverageResult

	query := r.db.WithContext(ctx).Unscoped().Table("partner_location_coverages plc").
		Where("plc.partner_id = ?", partnerID)

	if isActive {
		query = query.Where("plc.is_active = ?", true)
	}

	rows, err := query.Select("plc.partner_id, plc.postal_code, plc.postal_code_id, plc.zone_type").Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query partner coverage: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var result models.PartnerLocationCoverageResult
		if err := rows.Scan(&result.PartnerID, &result.PostalCode, &result.PostalCodeID, &result.ZoneType); err != nil {
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

	// Count total soft-deleted records
	err := r.db.WithContext(ctx).Unscoped().
		Model(&models.PartnerLocationCoverage{}).
		Where("is_deleted = ?", true).
		Count(&total).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count deleted partner location coverages: %w", err)
	}

	// Get paginated soft-deleted records
	err = r.db.WithContext(ctx).Unscoped().
		Where("is_deleted = ?", true).
		Offset(offset).
		Limit(limit).
		Find(&coverages).Error
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get deleted partner location coverages: %w", err)
	}

	return coverages, total, nil
}

// GetByLocationScopeAndID retrieves coverage records for a specific location (deprecated - use GetByPostalCode)
func (r *partnerLocationCoverageRepository) GetByLocationScopeAndID(ctx context.Context, locationScope string, locationID uint) ([]models.PartnerLocationCoverage, error) {
	// For backward compatibility, if locationScope is "POSTAL_CODE", treat locationID as postal code
	if strings.ToUpper(locationScope) == "POSTAL_CODE" {
		// This is a bit of a hack but maintains compatibility
		return r.GetByPostalCode(ctx, fmt.Sprintf("%d", locationID))
	}

	// For other location scopes, return empty result since the new schema only supports postal codes
	return []models.PartnerLocationCoverage{}, nil
}

// GetByLocationScopeAndIDWithDeleted retrieves coverage records for a specific location (includes soft-deleted records) (deprecated - use GetByPostalCodeWithDeleted)
func (r *partnerLocationCoverageRepository) GetByLocationScopeAndIDWithDeleted(ctx context.Context, locationScope string, locationID uint) ([]models.PartnerLocationCoverage, error) {
	// For backward compatibility, if locationScope is "POSTAL_CODE", treat locationID as postal code
	if strings.ToUpper(locationScope) == "POSTAL_CODE" {
		// This is a bit of a hack but maintains compatibility
		return r.GetByPostalCodeWithDeleted(ctx, fmt.Sprintf("%d", locationID))
	}

	// For other location scopes, return empty result since the new schema only supports postal codes
	return []models.PartnerLocationCoverage{}, nil
}

// GetPartnersForPostalCode retrieves all partners serving a specific postal code with zone priorities
func (r *partnerLocationCoverageRepository) GetPartnersForPostalCode(ctx context.Context, postalCode string, postalCodeID uint, zoneTypes []string) ([]models.PartnerLocationCoverageResult, error) {
	var results []models.PartnerLocationCoverageResult

	query := r.db.WithContext(ctx).Table("partner_location_coverages plc").
		Where("plc.postal_code = ? AND plc.is_active = ?", postalCode, true)

	if len(zoneTypes) > 0 {
		upperZoneTypes := make([]string, len(zoneTypes))
		for i, zt := range zoneTypes {
			upperZoneTypes[i] = strings.ToUpper(zt)
		}
		query = query.Where("plc.zone_type IN ?", upperZoneTypes)
	}

	// Order by zone type priority (PRIMARY, SECONDARY, BUFFER)
	query = query.Order("CASE plc.zone_type WHEN 'PRIMARY' THEN 1 WHEN 'SECONDARY' THEN 2 WHEN 'BUFFER' THEN 3 END")

	rows, err := query.Select("plc.partner_id, plc.postal_code, plc.postal_code_id, plc.zone_type").Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query partners for location: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var result models.PartnerLocationCoverageResult
		if err := rows.Scan(&result.PartnerID, &result.PostalCode, &result.PostalCodeID, &result.ZoneType); err != nil {
			return nil, fmt.Errorf("failed to scan partner result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetPartnersForPostalCodeWithDeleted retrieves all partners serving a specific postal code with zone priorities (includes soft-deleted records)
func (r *partnerLocationCoverageRepository) GetPartnersForPostalCodeWithDeleted(ctx context.Context, postalCode string, postalCodeID uint, zoneTypes []string) ([]models.PartnerLocationCoverageResult, error) {
	var results []models.PartnerLocationCoverageResult

	query := r.db.WithContext(ctx).Unscoped().Table("partner_location_coverages plc").
		Where("plc.postal_code = ?", postalCode)

	if len(zoneTypes) > 0 {
		upperZoneTypes := make([]string, len(zoneTypes))
		for i, zt := range zoneTypes {
			upperZoneTypes[i] = strings.ToUpper(zt)
		}
		query = query.Where("plc.zone_type IN ?", upperZoneTypes)
	}

	// Order by zone type priority (PRIMARY, SECONDARY, BUFFER)
	query = query.Order("CASE plc.zone_type WHEN 'PRIMARY' THEN 1 WHEN 'SECONDARY' THEN 2 WHEN 'BUFFER' THEN 3 END")

	rows, err := query.Select("plc.partner_id, plc.postal_code, plc.postal_code_id, plc.zone_type").Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query partners for location: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var result models.PartnerLocationCoverageResult
		if err := rows.Scan(&result.PartnerID, &result.PostalCode, &result.PostalCodeID, &result.ZoneType); err != nil {
			return nil, fmt.Errorf("failed to scan partner result: %w", err)
		}
		results = append(results, result)
	}

	return results, nil
}

// GetByPartnerCode retrieves all coverage records for a partner by partner code
func (r *partnerLocationCoverageRepository) GetByPartnerCode(ctx context.Context, partnerCode string) ([]models.PartnerLocationCoverage, error) {
	var coverages []models.PartnerLocationCoverage

	err := r.db.WithContext(ctx).
		Where("partner_code = ? AND is_active = ?", partnerCode, true).
		Find(&coverages).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get partner coverage by code: %w", err)
	}

	return coverages, nil
}
