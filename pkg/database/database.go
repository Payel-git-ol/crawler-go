package database

import (
	"Fyne-on/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	dsn := "host=localhost user=scraper password=scraper_pass dbname=scraper_postgres port=5432 sslmode=disable options='-c client_encoding=UTF8'"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Migrator().DropTable(&models.Issue{})
	db.Migrator().DropTable(&models.Repo{})
	db.AutoMigrate(&models.Issue{}, &models.Repo{}, &models.IssuesResponse{})
	return db
}
