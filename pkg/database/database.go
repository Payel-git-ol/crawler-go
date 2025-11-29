package database

import (
	"Fyne-on/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	dsn := "host=localhost user=scraper password=scraper_pass dbname=scraper_postgres port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.AutoMigrate(models.Issue{}, &models.Repo{})
	return db
}
