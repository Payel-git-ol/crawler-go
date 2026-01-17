package main

import (
	"Fyne-on/pkg/storage"
	"github.com/gofiber/fiber/v3"
)

func GetStats(c fiber.Ctx, storageService *storage.StorageService) error {
	stats := storageService.GetStats()
	return c.JSON(stats)
}

func GetStatsSummary(c fiber.Ctx, storageService *storage.StorageService) error {
	summary, err := storageService.GetCounts()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(summary)
}
