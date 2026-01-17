package main

import (
	"Fyne-on/pkg/parquet"
	"github.com/gofiber/fiber/v3"
)

var exporter = parquet.NewParquetExporter(DB)

func ExportIssues(c fiber.Ctx) error {
	outputPath := "./parquet_data/issues.parquet"
	count, err := exporter.ExportIssuesToParquet(outputPath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"status":    "success",
		"message":   "Issues exported to Parquet",
		"count":     count,
		"output":    outputPath,
		"format":    "parquet",
		"timestamp": c.Get("Date"),
	})
}

func ExportPullRequest(c fiber.Ctx) error {
	outputPath := "./parquet_data/pull_requests.parquet"
	count, err := exporter.ExportPullRequestsToParquet(outputPath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"status":    "success",
		"message":   "Pull requests exported to Parquet",
		"count":     count,
		"output":    outputPath,
		"format":    "parquet",
		"timestamp": c.Get("Date"),
	})
}

func ExportRepos(c fiber.Ctx) error {
	outputPath := "./parquet_data/repositories.parquet"
	count, err := exporter.ExportRepositoriesToParquet(outputPath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"status":    "success",
		"message":   "Repositories exported to Parquet",
		"count":     count,
		"output":    outputPath,
		"format":    "parquet",
		"timestamp": c.Get("Date"),
	})
}

func ExportAll(c fiber.Ctx) error {
	outputDir := "./parquet_data"
	results, err := exporter.ExportAllToParquet(outputDir)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"status":        "success",
		"message":       "All data exported to Parquet",
		"results":       results,
		"output_dir":    outputDir,
		"format":        "parquet",
		"total_records": results["issues"] + results["pull_requests"] + results["repositories"],
		"timestamp":     c.Get("Date"),
	})
}

func ExportAllJsonl(c fiber.Ctx) error {
	outputDir := "./jsonl_data"
	results, err := exporter.ExportAllToJSONL(outputDir)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"status":        "success",
		"message":       "All data exported to JSONL (perfect for LLM training)",
		"results":       results,
		"output_dir":    outputDir,
		"format":        "jsonl",
		"total_records": results["issues"] + results["pull_requests"] + results["repositories"],
		"note":          "JSONL format is better for LLM training - each line is a valid JSON object",
		"timestamp":     c.Get("Date"),
	})
}
