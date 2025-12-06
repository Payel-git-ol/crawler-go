package main

import (
	"Fyne-on/pkg/crawler"
	"Fyne-on/pkg/database"
	"Fyne-on/pkg/storage"
	"encoding/json" // added
	"log"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func main() {
	// Initialize Badger database
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize storage service
	storageService := storage.NewStorageService(db)

	// Initialize GitHub crawler
	githubCrawler := crawler.NewGithubCrawler(storageService)
	githubCrawler.SetMaxIterations(20000)
	githubCrawler.SetDelayMs(1000)

	// ADD: track current crawler config BEFORE usage in handlers
	currentCrawlerConfig := struct {
		StartUsername string
		MaxIterations int
		DelayMs       int
		TokenSet      bool
	}{
		StartUsername: "",
		MaxIterations: 20000,
		DelayMs:       1000,
		TokenSet:      false,
	}

	// Create Fiber app
	app := fiber.New()

	// Health check
	app.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Fyne-on crawler is running",
		})
	})

	// Get statistics (existing)
	app.Get("/stats", func(c fiber.Ctx) error {
		stats := storageService.GetStats()
		return c.JSON(stats)
	})

	// Add: compact JSON summary counts you requested
	app.Get("/stats/summary", func(c fiber.Ctx) error {
		summary, err := storageService.GetCounts()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.JSON(summary)
	})

	// Get all repositories
	app.Get("/repos", func(c fiber.Ctx) error {
		repos, err := storageService.GetAllRepos()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(repos)
	})

	// Get repository by owner and name
	app.Get("/repos/:owner/:name", func(c fiber.Ctx) error {
		owner := c.Params("owner")
		name := c.Params("name")

		repo, err := storageService.GetRepo(owner, name)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "repository not found"})
		}

		return c.JSON(repo)
	})

	// Get repository issues
	app.Get("/repos/:owner/:name/issues", func(c fiber.Ctx) error {
		owner := c.Params("owner")
		name := c.Params("name")
		repoID := owner + "/" + name

		issues, err := storageService.GetRepoIssues(repoID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(issues)
	})

	// Get repository pull requests
	app.Get("/repos/:owner/:name/prs", func(c fiber.Ctx) error {
		owner := c.Params("owner")
		name := c.Params("name")
		repoID := owner + "/" + name

		prs, err := storageService.GetRepoPullRequests(repoID)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(prs)
	})

	// Get all contacts
	app.Get("/contacts", func(c fiber.Ctx) error {
		contacts, err := storageService.GetAllContacts()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(contacts)
	})

	// Get contact by login
	app.Get("/contacts/:login", func(c fiber.Ctx) error {
		login := c.Params("login")

		contact, err := storageService.GetContact(login)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "contact not found"})
		}

		return c.JSON(contact)
	})

	// Start crawler (async)
	// Запуск краулера с параметрами
	type CrawlRequest struct {
		StartUsername string `json:"start_username"`
		MaxIterations int    `json:"max_iterations"`
		DelayMs       int    `json:"delay_ms"`
		GitHubToken   string `json:"github_token"`
		UsePlaywright bool   `json:"use_playwright"`
	}

	// Start crawler (fixed: manual JSON parsing + use CrawlStart)
	app.Post("/crawler/start", func(c fiber.Ctx) error {
		body := c.Body()

		type startReq struct {
			StartUsername string `json:"start_username"`
			MaxIterations int    `json:"max_iterations"`
			DelayMs       int    `json:"delay_ms"`
			GitHubToken   string `json:"github_token"`
		}
		var req startReq
		if err := json.Unmarshal(body, &req); err != nil {
			// try alternate key casing used in scripts
			type altReq struct {
				StartUsername string `json:"StartUsername"`
				MaxIter       int    `json:"MaxIter"`
				DelayMs       int    `json:"DelayMs"`
				GitHubToken   string `json:"GitHubToken"`
			}
			var alt altReq
			if err2 := json.Unmarshal(body, &alt); err2 != nil {
				return c.Status(400).JSON(fiber.Map{"error": "invalid JSON: " + err.Error()})
			}
			req.StartUsername = alt.StartUsername
			if alt.MaxIter > 0 {
				req.MaxIterations = alt.MaxIter
			}
			if alt.DelayMs >= 0 {
				req.DelayMs = alt.DelayMs
			}
			req.GitHubToken = alt.GitHubToken
		}

		// Apply crawler settings
		if req.GitHubToken != "" {
			githubCrawler.SetGitHubToken(req.GitHubToken)
			currentCrawlerConfig.TokenSet = true
		}
		if req.MaxIterations > 0 {
			githubCrawler.SetMaxIterations(req.MaxIterations)
			currentCrawlerConfig.MaxIterations = req.MaxIterations
		}
		if req.DelayMs >= 0 {
			githubCrawler.SetDelayMs(req.DelayMs)
			currentCrawlerConfig.DelayMs = req.DelayMs
		}
		if req.StartUsername == "" {
			req.StartUsername = "torvalds"
		}
		currentCrawlerConfig.StartUsername = req.StartUsername

		// Run crawler asynchronously
		go func(u string) {
			if err := githubCrawler.CrawlStart(u); err != nil {
				log.Printf("Crawler error: %v", err)
			}
		}(req.StartUsername)

		return c.JSON(fiber.Map{
			"message":        "Crawler started",
			"start_username": req.StartUsername,
			"max_iterations": currentCrawlerConfig.MaxIterations,
			"delay_ms":       currentCrawlerConfig.DelayMs,
		})
	})

	// Query repositories (filter by language, stars, etc.)
	app.Get("/repos/search", func(c fiber.Ctx) error {
		language := c.Query("language")
		minStars := c.Query("min_stars", "0")

		minStarsInt, _ := strconv.Atoi(minStars)

		repos, err := storageService.GetAllRepos()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		filtered := []interface{}{}
		for _, repo := range repos {
			if language != "" && repo.Language != language {
				continue
			}
			if repo.Stars < minStarsInt {
				continue
			}
			filtered = append(filtered, repo)
		}

		return c.JSON(filtered)
	})

	// Delete repository (cascade delete issues and PRs)
	app.Delete("/repos/:owner/:name", func(c fiber.Ctx) error {
		owner := c.Params("owner")
		name := c.Params("name")

		if err := storageService.DeleteRepo(owner, name); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{"message": "repository deleted"})
	})

	// INSERT: endpoints that were only present in the removed duplicate block
	app.Get("/crawler/config", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"start_username": currentCrawlerConfig.StartUsername,
			"max_iterations": currentCrawlerConfig.MaxIterations,
			"delay_ms":       currentCrawlerConfig.DelayMs,
			"token_set":      currentCrawlerConfig.TokenSet,
		})
	})

	app.Get("/api/routes", func(c fiber.Ctx) error {
		routes := []fiber.Map{
			{"method": "GET", "path": "/health", "description": "Health check"},
			{"method": "GET", "path": "/stats", "description": "Get database statistics"},
			{"method": "GET", "path": "/repos", "description": "Get all repositories"},
			{"method": "GET", "path": "/repos/:owner/:name", "description": "Get specific repository"},
			{"method": "GET", "path": "/repos/:owner/:name/issues", "description": "Get repository issues"},
			{"method": "GET", "path": "/repos/:owner/:name/prs", "description": "Get repository pull requests"},
			{"method": "GET", "path": "/repos/search", "description": "Search repositories (query: language, min_stars)"},
			{"method": "DELETE", "path": "/repos/:owner/:name", "description": "Delete repository"},
			{"method": "GET", "path": "/contacts", "description": "Get all contacts"},
			{"method": "GET", "path": "/contacts/:login", "description": "Get specific contact"},
			{"method": "POST", "path": "/crawler/start", "description": "Start crawler (body: start_username, max_iterations, delay_ms, github_token)"},
			{"method": "GET", "path": "/crawler/config", "description": "Get current crawler configuration"},
			{"method": "GET", "path": "/api/routes", "description": "List all available endpoints"},
		}
		return c.JSON(routes)
	})

	// Start server
	port := ":3000"
	log.Printf("Server started on %s", port)
	if err := app.Listen(port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
