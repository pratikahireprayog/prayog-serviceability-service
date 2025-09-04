package dtos

type CreateNearestHubLocationRequest struct {
	PostalCode  int      `json:"postal_code"`
	Address     *string  `json:"address,omitempty"`
	CentroidLat *float64 `json:"centroid_lat,omitempty"`
	CentroidLng *float64 `json:"centroid_lng,omitempty"`

	// Original city code field (renamed from hub_city_code)
	CityCode *string `json:"city_code,omitempty"`

	// Hub location fields (renamed from international_hub_*)
	HubPostalCode  *int     `json:"hub_postal_code,omitempty"`
	HubCentroidLat *float64 `json:"hub_centroid_lat,omitempty"`
	HubCentroidLng *float64 `json:"hub_centroid_lng,omitempty"`
	HubCityCode    *string  `json:"hub_city_code,omitempty"`

	// International hub flag
	IsInternationalHub *bool `json:"is_international_hub,omitempty"`

	// Hub contact and location fields
	HubContactPersonName  *string  `json:"hub_contact_person_name,omitempty"`
	HubContactPersonPhone *string  `json:"hub_contact_person_phone,omitempty"`
	HubContactPersonEmail *string  `json:"hub_contact_person_email,omitempty"`
	HubStreet             *string  `json:"hub_street,omitempty"`
	HubLandmark           *string  `json:"hub_landmark,omitempty"`
	HubCity               *string  `json:"hub_city,omitempty"`
	HubState              *string  `json:"hub_state,omitempty"`
	HubCountry            *string  `json:"hub_country,omitempty"`
	HubLat                *float64 `json:"hub_lat,omitempty"`
	HubLng                *float64 `json:"hub_lng,omitempty"`
}

type UpdateNearestHubLocationRequest struct {
	Address     *string  `json:"address,omitempty"`
	CentroidLat *float64 `json:"centroid_lat,omitempty"`
	CentroidLng *float64 `json:"centroid_lng,omitempty"`

	// Original city code field (renamed from hub_city_code)
	CityCode *string `json:"city_code,omitempty"`

	// Hub location fields (renamed from international_hub_*)
	HubPostalCode  *int     `json:"hub_postal_code,omitempty"`
	HubCentroidLat *float64 `json:"hub_centroid_lat,omitempty"`
	HubCentroidLng *float64 `json:"hub_centroid_lng,omitempty"`
	HubCityCode    *string  `json:"hub_city_code,omitempty"`

	// International hub flag
	IsInternationalHub *bool `json:"is_international_hub,omitempty"`

	// Hub contact and location fields
	HubContactPersonName  *string  `json:"hub_contact_person_name,omitempty"`
	HubContactPersonPhone *string  `json:"hub_contact_person_phone,omitempty"`
	HubContactPersonEmail *string  `json:"hub_contact_person_email,omitempty"`
	HubStreet             *string  `json:"hub_street,omitempty"`
	HubLandmark           *string  `json:"hub_landmark,omitempty"`
	HubCity               *string  `json:"hub_city,omitempty"`
	HubState              *string  `json:"hub_state,omitempty"`
	HubCountry            *string  `json:"hub_country,omitempty"`
	HubLat                *float64 `json:"hub_lat,omitempty"`
	HubLng                *float64 `json:"hub_lng,omitempty"`
}

type NearestHubLocationResponse struct {
	PostalCode  int      `json:"postal_code"`
	Address     *string  `json:"address,omitempty"`
	CentroidLat *float64 `json:"centroid_lat,omitempty"`
	CentroidLng *float64 `json:"centroid_lng,omitempty"`

	// Original city code field (renamed from hub_city_code)
	CityCode *string `json:"city_code,omitempty"`

	// Hub location fields (renamed from international_hub_*)
	HubPostalCode  *int     `json:"hub_postal_code,omitempty"`
	HubCentroidLat *float64 `json:"hub_centroid_lat,omitempty"`
	HubCentroidLng *float64 `json:"hub_centroid_lng,omitempty"`
	HubCityCode    *string  `json:"hub_city_code,omitempty"`

	// International hub flag
	IsInternationalHub *bool `json:"is_international_hub,omitempty"`

	// Hub contact and location fields
	HubContactPersonName  *string  `json:"hub_contact_person_name,omitempty"`
	HubContactPersonPhone *string  `json:"hub_contact_person_phone,omitempty"`
	HubContactPersonEmail *string  `json:"hub_contact_person_email,omitempty"`
	HubStreet             *string  `json:"hub_street,omitempty"`
	HubLandmark           *string  `json:"hub_landmark,omitempty"`
	HubCity               *string  `json:"hub_city,omitempty"`
	HubState              *string  `json:"hub_state,omitempty"`
	HubCountry            *string  `json:"hub_country,omitempty"`
	HubLat                *float64 `json:"hub_lat,omitempty"`
	HubLng                *float64 `json:"hub_lng,omitempty"`
}

type NearestHubLocationFilters struct {
	PostalCodes        []int    `json:"postal_codes,omitempty" query:"postal_code"`
	CityCodes          []string `json:"city_codes,omitempty" query:"city_code"`
	HubPostalCodes     []int    `json:"hub_postal_codes,omitempty" query:"hub_postal_code"`
	HubCityCodes       []string `json:"hub_city_codes,omitempty" query:"hub_city_code"`
	HubCities          []string `json:"hub_cities,omitempty" query:"hub_city"`
	HubStates          []string `json:"hub_states,omitempty" query:"hub_state"`
	HubCountries       []string `json:"hub_countries,omitempty" query:"hub_country"`
	IsInternationalHub *bool    `json:"is_international_hub,omitempty" query:"is_international_hub"`
}
