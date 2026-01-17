package main

import (
	"github.com/gofiber/fiber/v3"
)

func GetStats(c fiber.Ctx) error {
	if StorageService == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Storage service not initialized",
		})
	}

	stats := StorageService.GetStats()
	return c.JSON(stats)
}

func GetStatsSummary(c fiber.Ctx) error {
	if StorageService == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Storage service not initialized",
		})
	}

	summary, err := StorageService.GetCounts()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(summary)
}
