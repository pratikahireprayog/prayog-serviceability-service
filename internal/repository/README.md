# Repository Layer

This directory contains the repository layer implementations. The repository layer is responsible for data access and storage operations, implementing the repository interfaces defined in the domain layer.

## Structure

* `/gorm` - GORM-based repository implementations

## Responsibilities

* Implementing domain repository interfaces
* Database operations (CRUD)
* Data mapping between domain entities and database models
* Transaction management
* Query optimization
* Database-specific error handling

## Guidelines

* Repositories should implement interfaces defined in the domain layer
* Keep SQL queries and database-specific code isolated to this layer
* Handle database errors and map them to domain errors
* Use transactions when operations need to be atomic
* Implement proper data mapping to isolate domain model from database schema
* Document complex queries
* Write comprehensive tests with database mocks

## Example

```go
type LocationRepositoryImpl struct {
    db *gorm.DB
}

func NewLocationRepository(db *gorm.DB) domain.LocationRepository {
    return &LocationRepositoryImpl{db: db}
}

func (r *LocationRepositoryImpl) FindByID(ctx context.Context, id string) (*domain.Location, error) {
    var model Model
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, domain.ErrLocationNotFound
        }
        return nil, err
    }
    return mapModelToDomain(&model), nil
}
``` 