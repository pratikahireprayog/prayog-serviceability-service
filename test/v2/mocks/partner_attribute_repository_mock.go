package mocks

import (
	"context"
	"errors"

	"prayog-serviceability-service/internal/shared/models/v1"
	"prayog-serviceability-service/internal/shared/repositories/v1"

	"github.com/google/uuid"
)

// MockPartnerAttributeMapRepository implements the repositories.PartnerAttributeMapRepository interface for testing
type MockPartnerAttributeMapRepository struct {
	// Data storage
	PartnerAttributeMaps map[string]*models.PartnerAttributeMap
	PartnerInfos         map[string][]models.PartnerAttributeMap

	// Error configuration
	GetByIDError                   error
	GetByPartnerCodeError          error
	GetByAttributeCodeError        error
	GetPartnerInfoByAttributeError error
	GetAttributesByPartnerError    error

	// Call tracking
	GetByIDCalled                   bool
	GetByPartnerCodeCalled          bool
	GetByAttributeCodeCalled        bool
	GetPartnerInfoByAttributeCalled bool
	GetAttributesByPartnerCalled    bool

	// Last call parameters
	LastID            string
	LastPartnerCode   string
	LastAttributeCode string
	LastContext       context.Context
}

// NewMockPartnerAttributeMapRepository creates a new mock repository
func NewMockPartnerAttributeMapRepository() *MockPartnerAttributeMapRepository {
	return &MockPartnerAttributeMapRepository{
		PartnerAttributeMaps: make(map[string]*models.PartnerAttributeMap),
		PartnerInfos:         make(map[string][]models.PartnerAttributeMap),
	}
}

// GetByID mocks retrieving a partner attribute map by ID
func (m *MockPartnerAttributeMapRepository) GetByID(ctx context.Context, id string) (*models.PartnerAttributeMap, error) {
	m.GetByIDCalled = true
	m.LastID = id
	m.LastContext = ctx

	if m.GetByIDError != nil {
		return nil, m.GetByIDError
	}

	if mapping, exists := m.PartnerAttributeMaps[id]; exists {
		return mapping, nil
	}

	return nil, errors.New("partner attribute map not found")
}

// GetByIDWithDeleted mocks retrieving a partner attribute map by ID including deleted ones
func (m *MockPartnerAttributeMapRepository) GetByIDWithDeleted(ctx context.Context, id string) (*models.PartnerAttributeMap, error) {
	return m.GetByID(ctx, id) // Same implementation for mock
}

// GetByPartnerID mocks retrieving partner attribute maps by partner ID
func (m *MockPartnerAttributeMapRepository) GetByPartnerID(ctx context.Context, partnerID string) ([]models.PartnerAttributeMap, error) {
	m.LastContext = ctx

	if m.GetByPartnerCodeError != nil {
		return nil, m.GetByPartnerCodeError
	}

	if mappings, exists := m.PartnerInfos[partnerID]; exists {
		return mappings, nil
	}

	return []models.PartnerAttributeMap{}, nil
}

// GetByPartnerIDWithDeleted mocks retrieving partner attribute maps by partner ID including deleted ones
func (m *MockPartnerAttributeMapRepository) GetByPartnerIDWithDeleted(ctx context.Context, partnerID string) ([]models.PartnerAttributeMap, error) {
	return m.GetByPartnerID(ctx, partnerID) // Same implementation for mock
}

// GetByPartnerCode mocks retrieving partner attribute maps by partner code
func (m *MockPartnerAttributeMapRepository) GetByPartnerCode(ctx context.Context, partnerCode string) ([]models.PartnerAttributeMap, error) {
	m.GetByPartnerCodeCalled = true
	m.LastPartnerCode = partnerCode
	m.LastContext = ctx

	if m.GetByPartnerCodeError != nil {
		return nil, m.GetByPartnerCodeError
	}

	if mappings, exists := m.PartnerInfos[partnerCode]; exists {
		return mappings, nil
	}

	return []models.PartnerAttributeMap{}, nil
}

// GetByPartnerCodeWithDeleted mocks retrieving partner attribute maps by partner code including deleted ones
func (m *MockPartnerAttributeMapRepository) GetByPartnerCodeWithDeleted(ctx context.Context, partnerCode string) ([]models.PartnerAttributeMap, error) {
	return m.GetByPartnerCode(ctx, partnerCode) // Same implementation for mock
}

// GetByAttributeID mocks retrieving partner attribute maps by attribute ID
func (m *MockPartnerAttributeMapRepository) GetByAttributeID(ctx context.Context, attributeID string) ([]models.PartnerAttributeMap, error) {
	m.LastContext = ctx

	if m.GetByAttributeCodeError != nil {
		return nil, m.GetByAttributeCodeError
	}

	return []models.PartnerAttributeMap{}, nil
}

// GetByAttributeIDWithDeleted mocks retrieving partner attribute maps by attribute ID including deleted ones
func (m *MockPartnerAttributeMapRepository) GetByAttributeIDWithDeleted(ctx context.Context, attributeID string) ([]models.PartnerAttributeMap, error) {
	return m.GetByAttributeID(ctx, attributeID) // Same implementation for mock
}

// GetByAttributeCode mocks retrieving partner attribute maps by attribute code
func (m *MockPartnerAttributeMapRepository) GetByAttributeCode(ctx context.Context, attributeCode string) ([]models.PartnerAttributeMap, error) {
	m.GetByAttributeCodeCalled = true
	m.LastAttributeCode = attributeCode
	m.LastContext = ctx

	if m.GetByAttributeCodeError != nil {
		return nil, m.GetByAttributeCodeError
	}

	if mappings, exists := m.PartnerInfos[attributeCode]; exists {
		return mappings, nil
	}

	return []models.PartnerAttributeMap{}, nil
}

// GetByAttributeCodeWithDeleted mocks retrieving partner attribute maps by attribute code including deleted ones
func (m *MockPartnerAttributeMapRepository) GetByAttributeCodeWithDeleted(ctx context.Context, attributeCode string) ([]models.PartnerAttributeMap, error) {
	return m.GetByAttributeCode(ctx, attributeCode) // Same implementation for mock
}

// GetByPartnerAndAttribute mocks retrieving a partner attribute map by partner and attribute
func (m *MockPartnerAttributeMapRepository) GetByPartnerAndAttribute(ctx context.Context, partnerCode, attributeID string) (*models.PartnerAttributeMap, error) {
	m.LastContext = ctx
	m.LastPartnerCode = partnerCode

	key := partnerCode + ":" + attributeID
	if mapping, exists := m.PartnerAttributeMaps[key]; exists {
		return mapping, nil
	}

	return nil, errors.New("partner attribute map not found")
}

// GetAll mocks retrieving all partner attribute maps with pagination
func (m *MockPartnerAttributeMapRepository) GetAll(ctx context.Context, offset, limit int) ([]models.PartnerAttributeMap, int64, error) {
	m.LastContext = ctx

	var all []models.PartnerAttributeMap
	for _, mapping := range m.PartnerAttributeMaps {
		all = append(all, *mapping)
	}

	return all, int64(len(all)), nil
}

// GetAllWithDeleted mocks retrieving all partner attribute maps including deleted ones
func (m *MockPartnerAttributeMapRepository) GetAllWithDeleted(ctx context.Context, offset, limit int) ([]models.PartnerAttributeMap, int64, error) {
	return m.GetAll(ctx, offset, limit) // Same implementation for mock
}

// GetOnlyDeleted mocks retrieving only deleted partner attribute maps
func (m *MockPartnerAttributeMapRepository) GetOnlyDeleted(ctx context.Context, offset, limit int) ([]models.PartnerAttributeMap, int64, error) {
	m.LastContext = ctx
	return []models.PartnerAttributeMap{}, 0, nil
}

// GetWithFilters mocks retrieving partner attribute maps with filters
func (m *MockPartnerAttributeMapRepository) GetWithFilters(ctx context.Context, filters *repositories.PartnerAttributeMapFilters) ([]models.PartnerAttributeMap, int64, error) {
	m.LastContext = ctx

	var filtered []models.PartnerAttributeMap
	for _, mapping := range m.PartnerAttributeMaps {
		// Apply filters if provided
		if filters.PartnerCode != nil && *filters.PartnerCode != mapping.PartnerCode {
			continue
		}
		if filters.IsActive != nil && *filters.IsActive != mapping.IsActive {
			continue
		}

		filtered = append(filtered, *mapping)
	}

	return filtered, int64(len(filtered)), nil
}

// GetPartnerInfoByAttribute mocks retrieving partner info by attribute code
func (m *MockPartnerAttributeMapRepository) GetPartnerInfoByAttribute(ctx context.Context, attributeCode string) ([]models.PartnerAttributeMap, error) {
	m.GetPartnerInfoByAttributeCalled = true
	m.LastAttributeCode = attributeCode
	m.LastContext = ctx

	if m.GetPartnerInfoByAttributeError != nil {
		return nil, m.GetPartnerInfoByAttributeError
	}

	if mappings, exists := m.PartnerInfos[attributeCode]; exists {
		return mappings, nil
	}

	return []models.PartnerAttributeMap{}, nil
}

// GetAttributesByPartner mocks retrieving attributes by partner code
func (m *MockPartnerAttributeMapRepository) GetAttributesByPartner(ctx context.Context, partnerCode string) ([]models.Attribute, error) {
	m.GetAttributesByPartnerCalled = true
	m.LastPartnerCode = partnerCode
	m.LastContext = ctx

	if m.GetAttributesByPartnerError != nil {
		return nil, m.GetAttributesByPartnerError
	}

	return []models.Attribute{}, nil
}

// Create mocks creating a partner attribute map
func (m *MockPartnerAttributeMapRepository) Create(ctx context.Context, mapping *models.PartnerAttributeMap) error {
	m.LastContext = ctx

	// Generate new ID if not set
	if mapping.ID == (uuid.UUID{}) {
		mapping.ID = uuid.New()
	}

	m.PartnerAttributeMaps[mapping.ID.String()] = mapping
	return nil
}

// Update mocks updating a partner attribute map
func (m *MockPartnerAttributeMapRepository) Update(ctx context.Context, id string, mapping *models.PartnerAttributeMap) error {
	m.LastContext = ctx
	m.LastID = id

	if _, exists := m.PartnerAttributeMaps[id]; !exists {
		return errors.New("partner attribute map not found")
	}

	m.PartnerAttributeMaps[id] = mapping
	return nil
}

// Delete mocks deleting a partner attribute map
func (m *MockPartnerAttributeMapRepository) Delete(ctx context.Context, id string) error {
	m.LastContext = ctx
	m.LastID = id

	if _, exists := m.PartnerAttributeMaps[id]; !exists {
		return errors.New("partner attribute map not found")
	}

	delete(m.PartnerAttributeMaps, id)
	return nil
}

// Restore mocks restoring a partner attribute map
func (m *MockPartnerAttributeMapRepository) Restore(ctx context.Context, id string) error {
	m.LastContext = ctx
	m.LastID = id
	return nil // Mock implementation
}

// ForceDelete mocks force deleting a partner attribute map
func (m *MockPartnerAttributeMapRepository) ForceDelete(ctx context.Context, id string) error {
	return m.Delete(ctx, id)
}

// BulkCreate mocks bulk creating partner attribute maps
func (m *MockPartnerAttributeMapRepository) BulkCreate(ctx context.Context, mappings []models.PartnerAttributeMap) error {
	m.LastContext = ctx

	for i, mapping := range mappings {
		// Generate new ID if not set
		if mapping.ID == (uuid.UUID{}) {
			mappings[i].ID = uuid.New()
		}
		m.PartnerAttributeMaps[mappings[i].ID.String()] = &mappings[i]
	}

	return nil
}

// BulkDelete mocks bulk deleting partner attribute maps
func (m *MockPartnerAttributeMapRepository) BulkDelete(ctx context.Context, ids []uuid.UUID) error {
	m.LastContext = ctx

	for _, id := range ids {
		delete(m.PartnerAttributeMaps, id.String())
	}

	return nil
}

// GetStats mocks retrieving partner attribute stats
func (m *MockPartnerAttributeMapRepository) GetStats(ctx context.Context, partnerCode string) (*repositories.PartnerAttributeStats, error) {
	m.LastContext = ctx
	m.LastPartnerCode = partnerCode

	return &repositories.PartnerAttributeStats{
		PartnerCode:    partnerCode,
		TotalMappings:  int64(len(m.PartnerAttributeMaps)),
		ActiveMappings: int64(len(m.PartnerAttributeMaps)),
	}, nil
}

// SetPartnerAttributeMap sets a partner attribute map for testing
func (m *MockPartnerAttributeMapRepository) SetPartnerAttributeMap(id string, mapping *models.PartnerAttributeMap) {
	m.PartnerAttributeMaps[id] = mapping
}

// SetPartnerInfo sets partner info for testing
func (m *MockPartnerAttributeMapRepository) SetPartnerInfo(key string, mappings []models.PartnerAttributeMap) {
	m.PartnerInfos[key] = mappings
}

// SetGetByIDError sets the error to return from GetByID
func (m *MockPartnerAttributeMapRepository) SetGetByIDError(err error) {
	m.GetByIDError = err
}

// SetGetByPartnerCodeError sets the error to return from GetByPartnerCode
func (m *MockPartnerAttributeMapRepository) SetGetByPartnerCodeError(err error) {
	m.GetByPartnerCodeError = err
}

// SetGetByAttributeCodeError sets the error to return from GetByAttributeCode
func (m *MockPartnerAttributeMapRepository) SetGetByAttributeCodeError(err error) {
	m.GetByAttributeCodeError = err
}

// SetGetPartnerInfoByAttributeError sets the error to return from GetPartnerInfoByAttribute
func (m *MockPartnerAttributeMapRepository) SetGetPartnerInfoByAttributeError(err error) {
	m.GetPartnerInfoByAttributeError = err
}

// SetGetAttributesByPartnerError sets the error to return from GetAttributesByPartner
func (m *MockPartnerAttributeMapRepository) SetGetAttributesByPartnerError(err error) {
	m.GetAttributesByPartnerError = err
}

// Reset resets all call tracking
func (m *MockPartnerAttributeMapRepository) Reset() {
	m.GetByIDCalled = false
	m.GetByPartnerCodeCalled = false
	m.GetByAttributeCodeCalled = false
	m.GetPartnerInfoByAttributeCalled = false
	m.GetAttributesByPartnerCalled = false
	m.LastID = ""
	m.LastPartnerCode = ""
	m.LastAttributeCode = ""
	m.LastContext = nil
}

// GetPartnerCodesByAttribute mocks retrieving partner codes by attribute code
func (m *MockPartnerAttributeMapRepository) GetPartnerCodesByAttribute(ctx context.Context, attributeCode string) ([]string, error) {
	m.LastContext = ctx
	m.LastAttributeCode = attributeCode

	if m.GetByAttributeCodeError != nil {
		return nil, m.GetByAttributeCodeError
	}

	// Get mappings for this attribute
	mappings, exists := m.PartnerInfos[attributeCode]
	if !exists {
		return []string{}, nil
	}

	// Extract unique partner codes
	partnerCodes := make([]string, 0, len(mappings))
	seen := make(map[string]bool)

	for _, mapping := range mappings {
		if !seen[mapping.PartnerCode] {
			partnerCodes = append(partnerCodes, mapping.PartnerCode)
			seen[mapping.PartnerCode] = true
		}
	}

	return partnerCodes, nil
}

// SetupDefaultData sets up default test data
func (m *MockPartnerAttributeMapRepository) SetupDefaultData() {
	// Setup partner attribute mappings for international shipping
	internationalMappings := []models.PartnerAttributeMap{
		{
			ID:          uuid.New(),
			PartnerCode: "dhl",
			IsActive:    true,
		},
		{
			ID:          uuid.New(),
			PartnerCode: "shipyaari",
			IsActive:    true,
		},
	}

	m.SetPartnerInfo("international", internationalMappings)

	// Setup partner info for specific partner codes
	for _, mapping := range internationalMappings {
		m.SetPartnerInfo(mapping.PartnerCode, []models.PartnerAttributeMap{mapping})
	}
}

// Compile time check to ensure MockPartnerAttributeMapRepository implements the interface
var _ repositories.PartnerAttributeMapRepository = (*MockPartnerAttributeMapRepository)(nil)
