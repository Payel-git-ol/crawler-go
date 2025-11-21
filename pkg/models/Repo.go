package models

type Repo struct {
	Name           string `json:"name"`
	URL            string `json:"url"`
	HasOpenLicense bool   `json:"has_open_license"`
}
