package main

import (
	"Fyne-on/internal/GetIssues"
	"Fyne-on/internal/ResponseIssuesService"
	"Fyne-on/pkg/database"
	"Fyne-on/pkg/models"
	"github.com/gofiber/fiber/v3"
	"strconv"
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
		db.Find(&repos)
		return c.JSON(repos)
	})

	app.Post("/issues/:id/response", func(c fiber.Ctx) error {
		idStr := c.Params("id")
		issueID, err := strconv.Atoi(idStr)
		if err != nil || issueID == 0 {
			return c.Status(400).JSON(fiber.Map{"message": "Invalid issue ID"})
		}

		var body struct {
			Text string `json:"text"`
		}
		if err := c.Bind().JSON(&body); err != nil {
			return c.Status(400).JSON(fiber.Map{"message": "Invalid JSON"})
		}

		if err := ResponseIssuesService.AddResponse(db, uint(issueID), body.Text); err != nil {
			return c.Status(500).JSON(fiber.Map{"message": err.Error()})
		}

		return c.JSON(fiber.Map{
			"message": "Response saved",
			"issueId": issueID,
			"text":    body.Text,
		})
	})

	app.Listen(":3000")
}
