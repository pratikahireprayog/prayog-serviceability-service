package dtos

type CreateNearestHubLocationRequest struct {
	PostalCode                  int      `json:"postal_code"`
	Address                     *string  `json:"address,omitempty"`
	CentroidLat                 *float64 `json:"centroid_lat,omitempty"`
	CentroidLng                 *float64 `json:"centroid_lng,omitempty"`
	InternationalHubPostalCode  *int     `json:"international_hub_postal_code,omitempty"`
	InternationalHubCentroidLat *float64 `json:"international_hub_centroid_lat,omitempty"`
	InternationalHubCentroidLng *float64 `json:"international_hub_centroid_lng,omitempty"`
	InternationalHubCityCode    *string  `json:"international_hub_city_code,omitempty"`
	HubCityCode                 *string  `json:"hub_city_code,omitempty"`

	// New hub contact and location fields
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
	Address                     *string  `json:"address,omitempty"`
	CentroidLat                 *float64 `json:"centroid_lat,omitempty"`
	CentroidLng                 *float64 `json:"centroid_lng,omitempty"`
	InternationalHubPostalCode  *int     `json:"international_hub_postal_code,omitempty"`
	InternationalHubCentroidLat *float64 `json:"international_hub_centroid_lat,omitempty"`
	InternationalHubCentroidLng *float64 `json:"international_hub_centroid_lng,omitempty"`
	InternationalHubCityCode    *string  `json:"international_hub_city_code,omitempty"`
	HubCityCode                 *string  `json:"hub_city_code,omitempty"`

	// New hub contact and location fields
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
	PostalCode                  int      `json:"postal_code"`
	Address                     *string  `json:"address,omitempty"`
	CentroidLat                 *float64 `json:"centroid_lat,omitempty"`
	CentroidLng                 *float64 `json:"centroid_lng,omitempty"`
	InternationalHubPostalCode  *int     `json:"international_hub_postal_code,omitempty"`
	InternationalHubCentroidLat *float64 `json:"international_hub_centroid_lat,omitempty"`
	InternationalHubCentroidLng *float64 `json:"international_hub_centroid_lng,omitempty"`
	InternationalHubCityCode    *string  `json:"international_hub_city_code,omitempty"`
	HubCityCode                 *string  `json:"hub_city_code,omitempty"`

	// New hub contact and location fields
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
	PostalCodes                 []int    `json:"postal_codes,omitempty" query:"postal_code"`
	InternationalHubPostalCodes []int    `json:"international_hub_postal_codes,omitempty" query:"international_hub_postal_code"`
	InternationalHubCityCodes   []string `json:"international_hub_city_codes,omitempty" query:"international_hub_city_code"`
	HubCities                   []string `json:"hub_cities,omitempty" query:"hub_city"`
	HubStates                   []string `json:"hub_states,omitempty" query:"hub_state"`
	HubCountries                []string `json:"hub_countries,omitempty" query:"hub_country"`
}
