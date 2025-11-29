package main

import (
	"Fyne-on/internal/GetIssues"
	"Fyne-on/internal/ResponseIssuesService"
	"Fyne-on/pkg/database"
	"Fyne-on/pkg/models"
	"Fyne-on/pkg/search"
	"context"
	"github.com/gofiber/fiber/v3"
	"github.com/typesense/typesense-go/v4/typesense/api"
	"strconv"
)

func ptr(s string) *string { return &s }

func main() {
	db := database.InitDB()
	app := fiber.New()
	tsClient := search.InitTypesense()

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

	app.Get("/issues/search", func(c fiber.Ctx) error {
		q := c.Query("q")

		searchParams := &api.SearchCollectionParams{
			Q:       &q,
			QueryBy: ptr("title,state,url"),
		}

		result, err := tsClient.Collection("issues").Documents().Search(context.Background(), searchParams)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": err.Error()})
		}

		return c.JSON(result.Hits)
	})

	app.Listen(":3000")
}
