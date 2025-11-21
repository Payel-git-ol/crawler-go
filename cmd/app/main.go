package main

import (
	"Fyne-on/internal/GetIssues"
	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	app.Get("/issues", func(c fiber.Ctx) error {
		issues, err := GetIssues.FetchIssues()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": err.Error()})
		}
		return c.JSON(issues)
	})

	app.Get("/repos/:org", func(c fiber.Ctx) error {
		org := c.Params("org")
		repos, err := GetIssues.FetchRepo(org)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"message": err.Error()})
		}
		return c.JSON(repos)
	})
	app.Listen(":3000")
}
