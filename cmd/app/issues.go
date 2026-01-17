package main

import (
	"Fyne-on/pkg/models"
	"github.com/gofiber/fiber/v3"
	"strconv"
	"time"
)

func GetIssues(c fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "100"))
	offset := (page - 1) * limit

	type result struct {
		issues []models.Issue
		err    error
	}

	resultChan := make(chan result, 1)

	go func() {
		issues, err := StorageService.GetIssuesPage(offset, limit)
		resultChan <- result{issues, err}
	}()

	select {
	case result := <-resultChan:
		if result.err != nil {
			return c.Status(500).JSON(fiber.Map{})
		}

		return c.JSON(result.issues)
	case <-time.After(5 * time.Second):
		return c.Status(504).JSON(fiber.Map{"error": "Request timeout"})
	}
}
