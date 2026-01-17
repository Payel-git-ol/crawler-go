package main

import (
	crawler2 "Fyne-on/internal/core/crawler"
	"Fyne-on/pkg/database"
	"Fyne-on/pkg/storage"
	"github.com/gofiber/fiber/v3"
	"log"
)

var DB *database.BadgerDB
var StorageService *storage.StorageService

func main() {
	Db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer Db.Close()

	DB = Db

	storageService := storage.NewStorageService(Db)
	StorageService = storageService

	githubCrawler := crawler2.NewGithubCrawler(storageService)
	githubCrawler.SetMaxIterations(100)
	githubCrawler.SetDelayMs(5)

	app := fiber.New()

	app.Get("/health", Health)
	app.Get("/stats", GetStats)
	app.Get("/stats/summary", GetStatsSummary)
	app.Get("/repos", GetRepos)
	app.Get("/repos/:owner/:name", GetReposOwnerName)
	app.Get("/issues", GetIssues)
	app.Get("/repos/:owner/:name/issues", GetReposOwnerNameIssues)
	app.Get("/repos/:owner/:name/prs", GetReposOwnerNamePrs)
	app.Get("/contacts", GetContacts)
	app.Get("/contacts/:login", GetContactsLogin)
	app.Post("/crawler/start", CrawlerStart)
	app.Get("/repos/search", ReposSearch)
	app.Delete("/repos/:owner/:name", DeleteReposOwnerName)
	app.Get("/crawler/config", CrawlerConfig)
	app.Get("/api/routes", GetRoutes)
	app.Post("/export/issues", ExportIssues)
	app.Post("/export/pull-requests", ExportPullRequest)
	app.Post("/export/repositories", ExportRepos)
	app.Post("/export/all", ExportAll)
	app.Post("/export/all-jsonl", ExportAllJsonl)

	port := ":3000"
	log.Printf("Server started on %s", port)
	if err := app.Listen(":3000"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
