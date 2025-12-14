package main

import (
	"Fyne-on/pkg/crawler"
	"Fyne-on/pkg/database"
	"Fyne-on/pkg/storage"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

func main() {
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	storageService := storage.NewStorageService(db)

	githubCrawler := crawler.NewGithubCrawler(storageService)
	githubCrawler.SetMaxIterations(100)
	githubCrawler.SetDelayMs(5)

	currentCrawlerConfig := struct {
		StartUsername string
		MaxIterations int
		DelayMs       int
		TokenSet      bool
		UsePlaywright bool
	}{
		StartUsername: "",
		MaxIterations: 20000,
		DelayMs:       1000,
		TokenSet:      false,
		UsePlaywright: false,
	}

	app := fiber.New()

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

	// Get all repositories (add optional issues_count via ?include_issues=count|true|1)
	app.Get("/repos", func(c fiber.Ctx) error {
		includeIssues := c.Query("include_issues")
		includeCount := includeIssues == "count" || includeIssues == "true" || includeIssues == "1"

		// NEW: optional expansion of fields
		expandQ := c.Query("expand")
		expand := expandQ == "1" || expandQ == "true" || expandQ == "full"

		repos, err := storageService.GetAllRepos()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		result := make([]fiber.Map, 0, len(repos))
		for _, repo := range repos {
			hash := repo.Hash
			if hash == "" {
				h := sha256.Sum256([]byte(repo.Owner + "/" + repo.Name))
				hash = hex.EncodeToString(h[:])
			}

			item := fiber.Map{
				"hash":     hash,
				"owner":    repo.Owner,
				"name":     repo.Name,
				"language": repo.Language,
				"url":      repo.URL,
			}

			// NEW: include additional fields if requested
			if expand {
				item["url"] = repo.URL
				item["description"] = repo.Description
				item["stars"] = repo.Stars
				item["license"] = repo.License
				item["has_open_license"] = repo.HasOpenLicense
				item["updated_at"] = repo.UpdatedAt
				item["createdAt"] = repo.CreatedAt
			}

			if includeCount {
				issues, err := storageService.GetRepoIssues(repo.Owner + "/" + repo.Name)
				if err == nil {
					item["issues_count"] = len(issues)
				} else {
					item["issues_count"] = 0
					log.Printf("failed to load issues for %s/%s: %v", repo.Owner, repo.Name, err)
				}
			}

			result = append(result, item)
		}
		return c.JSON(result)
	})

	// Get repository by owner and name
	app.Get("/repos/:owner/:name", func(c fiber.Ctx) error {
		owner := c.Params("owner")
		name := c.Params("name")

		// NEW: optional expansion of fields
		expandQ := c.Query("expand")
		expand := expandQ == "1" || expandQ == "true" || expandQ == "full"

		repo, err := storageService.GetRepo(owner, name)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "repository not found"})
		}

		hash := repo.Hash
		if hash == "" {
			h := sha256.Sum256([]byte(owner + "/" + name))
			hash = hex.EncodeToString(h[:])
		}

		if !expand {
			return c.JSON(fiber.Map{
				"hash":     hash,
				"owner":    owner,
				"name":     name,
				"language": repo.Language,
				"url":      repo.URL,
			})
		}

		// NEW: expanded shape if requested
		return c.JSON(fiber.Map{
			"hash":             hash,
			"owner":            owner,
			"name":             name,
			"language":         repo.Language,
			"url":              repo.URL,
			"description":      repo.Description,
			"stars":            repo.Stars,
			"license":          repo.License,
			"has_open_license": repo.HasOpenLicense,
			"updated_at":       repo.UpdatedAt,
			"createdAt":        repo.CreatedAt,
		})
	})

	app.Get("/issues", func(c fiber.Ctx) error {
		page, _ := strconv.Atoi(c.Query("page", "1"))
		limit, _ := strconv.Atoi(c.Query("limit", "100"))
		offset := (page - 1) * limit

		issues, err := storageService.GetIssuesPage(limit, offset)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(issues)
	})

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

	app.Get("/contacts", func(c fiber.Ctx) error {
		contacts, err := storageService.GetAllContacts()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		result := make([]fiber.Map, 0, len(contacts))
		for _, ct := range contacts {
			result = append(result, fiber.Map{
				"login": ct.Login,
				"hash":  ct.Hash,
			})
		}
		return c.JSON(result)
	})

	app.Get("/contacts/:login", func(c fiber.Ctx) error {
		login := c.Params("login")

		contact, err := storageService.GetContact(login)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "contact not found"})
		}

		return c.JSON(fiber.Map{
			"login": contact.Login,
			"hash":  contact.Hash,
		})
	})

	type CrawlRequest struct {
		StartUsernames []string `json:"start_usernames"`
		MaxIterations  int      `json:"max_iterations"`
		DelayMs        int      `json:"delay_ms"`
		GitHubToken    string   `json:"github_token"`
		UsePlaywright  bool     `json:"use_playwright"`
	}

	// Start crawler (fixed: manual JSON parsing + use CrawlStart)
	app.Post("/crawler/start", func(c fiber.Ctx) error {
		body := c.Body()

		type startReq struct {
			StartUsername string `json:"start_username"`
			MaxIterations int    `json:"max_iterations"`
			DelayMs       int    `json:"delay_ms"`
			GitHubToken   string `json:"github_token"`
			UsePlaywright bool   `json:"use_playwright"`
		}
		var req CrawlRequest
		if err := json.Unmarshal(body, &req); err != nil {
			// try alternate key casing used in scripts
			type altReq struct {
				StartUsername string `json:"StartUsername"`
				MaxIter       int    `json:"MaxIter"`
				DelayMs       int    `json:"DelayMs"`
				GitHubToken   string `json:"GitHubToken"`
				UsePlaywright bool   `json:"UsePlaywright"`
			}
			var alt altReq
			if err2 := json.Unmarshal(body, &alt); err2 != nil {
				return c.Status(400).JSON(fiber.Map{"error": "invalid JSON: " + err.Error()})
			}
			req.StartUsernames = []string{alt.StartUsername}
			if alt.MaxIter > 0 {
				req.MaxIterations = alt.MaxIter
			}
			if alt.DelayMs >= 0 {
				req.DelayMs = alt.DelayMs
			}
			req.GitHubToken = alt.GitHubToken
			req.UsePlaywright = alt.UsePlaywright
		}

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
		currentCrawlerConfig.UsePlaywright = req.UsePlaywright

		if len(req.StartUsernames) == 0 {
			req.StartUsernames = []string{"microsoft"}
		}

		if req.UsePlaywright {
			go func(orgs []string) {
				if err := githubCrawler.CrawlStartOrgsHTML(orgs); err != nil {
					log.Printf("HTML crawler error: %v", err)
				}
			}(req.StartUsernames)
		} else {
			for _, user := range req.StartUsernames {
				go func(u string) {
					if err := githubCrawler.CrawlStart(u); err != nil {
						log.Printf("Crawler error for %s: %v", u, err)
					}
				}(user)
			}
		}

		return c.JSON(fiber.Map{
			"message":        "Crawler started (API mode)",
			"start_username": req.StartUsernames,
			"max_iterations": currentCrawlerConfig.MaxIterations,
			"delay_ms":       currentCrawlerConfig.DelayMs,
			"use_playwright": currentCrawlerConfig.UsePlaywright,
		})
	})

	app.Get("/repos/search", func(c fiber.Ctx) error {
		language := c.Query("language")

		repos, err := storageService.GetAllRepos()
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		filtered := []fiber.Map{}
		for _, repo := range repos {
			if language != "" && repo.Language != language {
				continue
			}
			hash := repo.Hash
			if hash == "" {
				h := sha256.Sum256([]byte(repo.Owner + "/" + repo.Name))
				hash = hex.EncodeToString(h[:])
			}
			filtered = append(filtered, fiber.Map{
				"hash":     hash,
				"owner":    repo.Owner,
				"name":     repo.Name,
				"language": repo.Language,
			})
		}

		return c.JSON(filtered)
	})

	app.Delete("/repos/:owner/:name", func(c fiber.Ctx) error {
		owner := c.Params("owner")
		name := c.Params("name")

		if err := storageService.DeleteRepo(owner, name); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{"message": "repository deleted"})
	})

	app.Get("/crawler/config", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"start_username": currentCrawlerConfig.StartUsername,
			"max_iterations": currentCrawlerConfig.MaxIterations,
			"delay_ms":       currentCrawlerConfig.DelayMs,
			"token_set":      currentCrawlerConfig.TokenSet,
			"use_playwright": currentCrawlerConfig.UsePlaywright,
		})
	})

	app.Get("/api/routes", func(c fiber.Ctx) error {
		routes := []fiber.Map{
			{"method": "GET", "path": "/health", "description": "Health check"},
			{"method": "GET", "path": "/stats", "description": "Get database statistics"},
			{"method": "GET", "path": "/repos", "description": "Get all repositories (query: include_issues=count)"},
			{"method": "GET", "path": "/repos/:owner/:name", "description": "Get specific repository"},
			{"method": "GET", "path": "/repos/:owner/:name/issues", "description": "Get repository issues"},
			{"method": "GET", "path": "/repos/:owner/:name/prs", "description": "Get repository pull requests"},
			{"method": "GET", "path": "/repos/search", "description": "Search repositories (query: language)"},
			{"method": "DELETE", "path": "/repos/:owner/:name", "description": "Delete repository"},
			{"method": "GET", "path": "/contacts", "description": "Get all contacts"},
			{"method": "GET", "path": "/contacts/:login", "description": "Get specific contact"},
			{"method": "POST", "path": "/crawler/start", "description": "Start crawler (HTML mode; body: start_username, max_iterations, delay_ms, github_token, use_playwright)"},
			{"method": "GET", "path": "/crawler/config", "description": "Get current crawler configuration"},
			{"method": "GET", "path": "/issues", "description": "Get all issues"},
			{"method": "GET", "path": "/api/routes", "description": "List all available endpoints"},
		}
		return c.JSON(routes)
	})

	port := ":3000"
	log.Printf("Server started on %s", port)
	if err := app.Listen(port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
