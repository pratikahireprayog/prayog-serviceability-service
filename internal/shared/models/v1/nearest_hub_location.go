package models

// NearestHubLocation represents a row in the nearest_hub_locations table
// All fields except postal_code are nullable

type NearestHubLocation struct {
	PostalCode                  int      `gorm:"column:postal_code;primaryKey"`
	Address                     *string  `gorm:"column:address"`
	CentroidLat                 *float64 `gorm:"column:centroid_lat"`
	CentroidLng                 *float64 `gorm:"column:centroid_lng"`
	InternationalHubPostalCode  *int     `gorm:"column:international_hub_postal_code"`
	InternationalHubAddress     *string  `gorm:"column:international_hub_address"`
	InternationalHubCentroidLat *float64 `gorm:"column:international_hub_centroid_lat"`
	InternationalHubCentroidLng *float64 `gorm:"column:international_hub_centroid_lng"`
	InternationalHubCityCode    *string  `gorm:"column:international_hub_city_code"`
	HubCityCode                 *string  `gorm:"column:hub_city_code"`
}

func (NearestHubLocation) TableName() string {
	return "nearest_hub_locations"
}
