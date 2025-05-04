package usecase

// UseCaseFactory provides access to all use cases
type UseCaseFactory struct {
	serviceabilityService ServiceabilityService
	adminService          AdminService
}

// NewUseCaseFactory creates a new use case factory
func NewUseCaseFactory() *UseCaseFactory {
	return &UseCaseFactory{
		serviceabilityService: NewServiceabilityService(),
		adminService:          NewAdminService(),
	}
}

// ServiceabilityService returns the serviceability service
func (f *UseCaseFactory) ServiceabilityService() ServiceabilityService {
	return f.serviceabilityService
}

// AdminService returns the admin service
func (f *UseCaseFactory) AdminService() AdminService {
	return f.adminService
}
