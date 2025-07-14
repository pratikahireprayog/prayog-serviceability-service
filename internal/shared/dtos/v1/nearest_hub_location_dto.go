package dtos

type CreateNearestHubLocationRequest struct {
	PostalCode                  int      `json:"postal_code"`
	Address                     *string  `json:"address,omitempty"`
	CentroidLat                 *float64 `json:"centroid_lat,omitempty"`
	CentroidLng                 *float64 `json:"centroid_lng,omitempty"`
	InternationalHubPostalCode  *int     `json:"international_hub_postal_code,omitempty"`
	InternationalHubAddress     *string  `json:"international_hub_address,omitempty"`
	InternationalHubCentroidLat *float64 `json:"international_hub_centroid_lat,omitempty"`
	InternationalHubCentroidLng *float64 `json:"international_hub_centroid_lng,omitempty"`
	InternationalHubCityCode    *string  `json:"international_hub_city_code,omitempty"`
	HubCityCode                 *string  `json:"hub_city_code,omitempty"`
}

type UpdateNearestHubLocationRequest struct {
	Address                     *string  `json:"address,omitempty"`
	CentroidLat                 *float64 `json:"centroid_lat,omitempty"`
	CentroidLng                 *float64 `json:"centroid_lng,omitempty"`
	InternationalHubPostalCode  *int     `json:"international_hub_postal_code,omitempty"`
	InternationalHubAddress     *string  `json:"international_hub_address,omitempty"`
	InternationalHubCentroidLat *float64 `json:"international_hub_centroid_lat,omitempty"`
	InternationalHubCentroidLng *float64 `json:"international_hub_centroid_lng,omitempty"`
	InternationalHubCityCode    *string  `json:"international_hub_city_code,omitempty"`
	HubCityCode                 *string  `json:"hub_city_code,omitempty"`
}

type NearestHubLocationResponse struct {
	PostalCode                  int      `json:"postal_code"`
	Address                     *string  `json:"address,omitempty"`
	CentroidLat                 *float64 `json:"centroid_lat,omitempty"`
	CentroidLng                 *float64 `json:"centroid_lng,omitempty"`
	InternationalHubPostalCode  *int     `json:"international_hub_postal_code,omitempty"`
	InternationalHubAddress     *string  `json:"international_hub_address,omitempty"`
	InternationalHubCentroidLat *float64 `json:"international_hub_centroid_lat,omitempty"`
	InternationalHubCentroidLng *float64 `json:"international_hub_centroid_lng,omitempty"`
	InternationalHubCityCode    *string  `json:"international_hub_city_code,omitempty"`
	HubCityCode                 *string  `json:"hub_city_code,omitempty"`
}

type NearestHubLocationFilters struct {
	PostalCodes                 []int    `json:"postal_codes,omitempty" query:"postal_code"`
	InternationalHubPostalCodes []int    `json:"international_hub_postal_codes,omitempty" query:"international_hub_postal_code"`
	InternationalHubCityCodes   []string `json:"international_hub_city_codes,omitempty" query:"international_hub_city_code"`
}
