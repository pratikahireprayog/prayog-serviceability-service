package models

import (
	"errors"

	"gorm.io/gorm"
)

// SoftDeleteQueryHelper provides utility functions for soft delete operations
type SoftDeleteQueryHelper struct {
	db *gorm.DB
}

// NewSoftDeleteQueryHelper creates a new instance of SoftDeleteQueryHelper
func NewSoftDeleteQueryHelper(db *gorm.DB) *SoftDeleteQueryHelper {
	return &SoftDeleteQueryHelper{db: db}
}

// WithDeletedRecords returns a GORM DB instance that includes soft-deleted records
// This bypasses the default scope that filters out soft-deleted records
func (h *SoftDeleteQueryHelper) WithDeletedRecords() *gorm.DB {
	return h.db.Unscoped()
}

// OnlyDeletedRecords returns a GORM DB instance that shows only soft-deleted records
func (h *SoftDeleteQueryHelper) OnlyDeletedRecords() *gorm.DB {
	return h.db.Unscoped().Where("is_deleted = ?", true)
}

// OnlyActiveRecords returns a GORM DB instance that shows only non-deleted records
// This is equivalent to the default scope but can be used explicitly
func (h *SoftDeleteQueryHelper) OnlyActiveRecords() *gorm.DB {
	return h.db.Where("is_deleted = ?", false)
}

// BulkSoftDelete performs soft delete operation on multiple records
// It accepts a slice of models that implement SoftDeleteModel interface
func (h *SoftDeleteQueryHelper) BulkSoftDelete(models []SoftDeleteModel) error {
	if len(models) == 0 {
		return errors.New("no models provided for bulk soft delete")
	}

	tx := h.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for _, model := range models {
		if err := model.SoftDelete(tx); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// BulkRestore performs restore operation on multiple soft-deleted records
// It accepts a slice of models that implement SoftDeleteModel interface
func (h *SoftDeleteQueryHelper) BulkRestore(models []SoftDeleteModel) error {
	if len(models) == 0 {
		return errors.New("no models provided for bulk restore")
	}

	tx := h.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for _, model := range models {
		if err := model.Restore(tx); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// CountDeleted returns the count of soft-deleted records for a given model type
func (h *SoftDeleteQueryHelper) CountDeleted(model interface{}) (int64, error) {
	var count int64
	err := h.db.Unscoped().Model(model).Where("is_deleted = ?", true).Count(&count).Error
	return count, err
}

// CountActive returns the count of active (non-deleted) records for a given model type
func (h *SoftDeleteQueryHelper) CountActive(model interface{}) (int64, error) {
	var count int64
	err := h.db.Model(model).Where("is_deleted = ?", false).Count(&count).Error
	return count, err
}

// FindDeletedByID finds a soft-deleted record by its ID
func (h *SoftDeleteQueryHelper) FindDeletedByID(model interface{}, id interface{}) error {
	return h.db.Unscoped().Where("is_deleted = ?", true).First(model, id).Error
}

// HardDelete permanently removes a soft-deleted record from the database
// WARNING: This operation is irreversible
func (h *SoftDeleteQueryHelper) HardDelete(model interface{}) error {
	return h.db.Unscoped().Delete(model).Error
}

// CleanupOldDeletedRecords permanently removes soft-deleted records older than specified days
// WARNING: This operation is irreversible
func (h *SoftDeleteQueryHelper) CleanupOldDeletedRecords(model interface{}, olderThanDays int) error {
	return h.db.Unscoped().
		Where("is_deleted = ? AND deleted_at < NOW() - INTERVAL ? DAY", true, olderThanDays).
		Delete(model).Error
}

// Global helper functions that can be used without creating a helper instance

// WithDeleted returns a GORM DB instance that includes soft-deleted records
func WithDeleted(db *gorm.DB) *gorm.DB {
	return db.Unscoped()
}

// OnlyDeleted returns a GORM DB instance that shows only soft-deleted records
func OnlyDeleted(db *gorm.DB) *gorm.DB {
	return db.Unscoped().Where("is_deleted = ?", true)
}

// OnlyActive returns a GORM DB instance that shows only non-deleted records
func OnlyActive(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", false)
}

// IsSoftDeleted checks if a model implementing SoftDeleteModel is soft-deleted
func IsSoftDeleted(model SoftDeleteModel) bool {
	return model.IsDeletedRecord()
}
