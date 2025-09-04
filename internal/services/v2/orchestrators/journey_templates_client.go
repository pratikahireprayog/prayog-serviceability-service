package orchestrators

import "context"

// JourneyTemplatesClient defines the minimal contract to fetch templates by parcel category
type JourneyTemplatesClient interface {
    GetTemplates(ctx context.Context, parcelCategory string) (*TemplatesResponse, error)
}

// TemplatesResponse represents the journey templates API response
type TemplatesResponse struct {
    Success bool       `json:"success"`
    Data    []Template `json:"data"`
}

// Template represents a single template
type Template struct {
    Code string `json:"code"`
}



