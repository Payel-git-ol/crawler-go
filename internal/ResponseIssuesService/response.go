package ResponseIssuesService

import (
	"Fyne-on/pkg/models"
	"gorm.io/gorm"
)

func AddResponse(db *gorm.DB, issueID uint, text string) error {
	response := models.IssuesResponse{
		IssueID: issueID,
		Text:    text,
	}
	return db.Create(&response).Error
}
