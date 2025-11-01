package models

// NaqelLocation represents a Naqel serviceability location entry
type NaqelLocation struct {
	CountryCode  string `gorm:"column:country_code;type:varchar(2);index"`
	LocationEn   string `gorm:"column:location_en;type:varchar(255)"`
	CityCode     string `gorm:"column:city_code;type:varchar(20);index"`
	StationCode  string `gorm:"column:station_code;type:varchar(20)"`
	IsServiceable bool  `gorm:"column:is_serviceable;type:boolean;default:true"`
}

// TableName returns the table name for NaqelLocation
func (NaqelLocation) TableName() string {
	return "naqel_cities"
}

