package ResponseIssuesService

import (
	"Fyne-on/pkg/models"
	"gorm.io/gorm"
)

func DeleteIssue(db *gorm.DB, issueID uint) error {
	if err := db.Where("issue_id = ?", issueID).Delete(&models.IssuesResponse{}).Error; err != nil {
		return err
	}
	if err := db.Delete(&models.Issue{}, issueID).Error; err != nil {
		return err
	}
	return nil
}
