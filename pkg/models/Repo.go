package models

type Repo struct {
	ID             uint   `gorm:"primary_key"`
	Name           string `json:"name"`
	URL            string `json:"url"`
	HasOpenLicense bool   `json:"has_open_license"`
}
