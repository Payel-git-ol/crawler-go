package main

import (
	"Fyne-on/internal/GetIssues"
	"Fyne-on/pkg/database"
	"Fyne-on/pkg/models"
	"github.com/gofiber/fiber/v3"
)

func main() {
	db := database.InitDB()
	app := fiber.New()

	_, err := GetIssues.FetchIssues(db)
	if err != nil {
		panic(err)
	}

	app.Get("/issues", func(c fiber.Ctx) error {
		var issues []models.Issue
		db.Find(&issues)
		return c.JSON(issues)
	})

	app.Get("/repos/:org", func(c fiber.Ctx) error {
		org := c.Params("org")
		_, err := GetIssues.FetchRepo(org, db)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": err.Error()})
		}

		var repos []models.Repo
		db.Find(&repos) // читаем из базы
		return c.JSON(repos)
	})

	app.Listen(":3000")
}
