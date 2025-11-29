package models

type Issue struct {
	ID        uint             `gorm:"primaryKey"`
	Title     string           `json:"title"`
	URL       string           `json:"url"`
	State     string           `json:"state"`
	Responses []IssuesResponse `gorm:"foreignKey:IssueID;constraint:OnDelete:CASCADE" json:"responses"`
}
