package usecase

import (
	"time"

	"github.com/prayog/serviceability/internal/domain"
)

// TimeRuleService defines operations for time-based rule validation
type TimeRuleService interface {
	// IsRuleActive checks if a rule is active at the specified time
	IsRuleActive(rule *domain.ServiceAvailability, timePoint time.Time) bool

	// FilterActiveRules filters a list of rules to only those active at the specified time
	FilterActiveRules(rules []*domain.ServiceAvailability, timePoint time.Time) []*domain.ServiceAvailability

	// GetMostRecentActiveRule returns the most recently activated rule among active rules
	GetMostRecentActiveRule(rules []*domain.ServiceAvailability, timePoint time.Time) *domain.ServiceAvailability

	// GetUpcomingRuleChanges returns future rule changes that are scheduled to take effect
	GetUpcomingRuleChanges(rules []*domain.ServiceAvailability, from time.Time) []*domain.ServiceAvailability
}

// timeRuleService implements the TimeRuleService interface
type timeRuleService struct{}

// NewTimeRuleService creates a new time rule service
func NewTimeRuleService() TimeRuleService {
	return &timeRuleService{}
}

// IsRuleActive checks if a rule is active at the specified time
func (s *timeRuleService) IsRuleActive(rule *domain.ServiceAvailability, timePoint time.Time) bool {
	// Check if the time point is within the effective range
	return rule.EffectiveFrom.Before(timePoint) && rule.EffectiveTo.After(timePoint)
}

// FilterActiveRules filters a list of rules to only those active at the specified time
func (s *timeRuleService) FilterActiveRules(rules []*domain.ServiceAvailability, timePoint time.Time) []*domain.ServiceAvailability {
	var activeRules []*domain.ServiceAvailability

	for _, rule := range rules {
		if s.IsRuleActive(rule, timePoint) {
			activeRules = append(activeRules, rule)
		}
	}

	return activeRules
}

// GetMostRecentActiveRule returns the most recently activated rule among active rules
func (s *timeRuleService) GetMostRecentActiveRule(rules []*domain.ServiceAvailability, timePoint time.Time) *domain.ServiceAvailability {
	var mostRecentRule *domain.ServiceAvailability
	var mostRecentTime time.Time

	// Filter to active rules first
	activeRules := s.FilterActiveRules(rules, timePoint)

	// Find the most recently activated rule
	for _, rule := range activeRules {
		if mostRecentRule == nil || rule.EffectiveFrom.After(mostRecentTime) {
			mostRecentRule = rule
			mostRecentTime = rule.EffectiveFrom
		}
	}

	return mostRecentRule
}

// GetUpcomingRuleChanges returns future rule changes that are scheduled to take effect
func (s *timeRuleService) GetUpcomingRuleChanges(rules []*domain.ServiceAvailability, from time.Time) []*domain.ServiceAvailability {
	var upcomingRules []*domain.ServiceAvailability

	for _, rule := range rules {
		// If the rule starts after the given time point, it's an upcoming change
		if rule.EffectiveFrom.After(from) {
			upcomingRules = append(upcomingRules, rule)
		}
	}

	return upcomingRules
}
