package models

type IssuesResponse struct {
	ID      uint   `gorm:"primaryKey"`
	IssueID uint   `gorm:"index"`
	Text    string `json:"text"`
}
