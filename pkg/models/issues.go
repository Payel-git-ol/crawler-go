package models

type Issue struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	State string `json:"state"`
}
