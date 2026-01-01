package main

import (
	crawler2 "Fyne-on/internal/core/crawler"
	"Fyne-on/pkg/database"
	"Fyne-on/pkg/models"
	"Fyne-on/pkg/parquet"
	"Fyne-on/pkg/storage"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func getLatestJobDate(jobs []models.TelegramJob) string {
	if len(jobs) == 0 {
		return "N/A"
	}
	latest := jobs[0].Timestamp
	for _, job := range jobs {
		if job.Timestamp.After(latest) {
			latest = job.Timestamp
		}
	}
	return latest.Format("2006-01-02 15:04:05")
}

func getAveragePayment(jobs []models.TelegramJob) string {
	if len(jobs) == 0 {
		return "N/A"
	}
	return "Рассчитывается"
}

func main() {
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	storageService := storage.NewStorageService(db)

	githubCrawler := crawler2.NewGithubCrawler(storageService)
	githubCrawler.SetMaxIterations(100)
	githubCrawler.SetDelayMs(5)

	telegramJobCrawler := crawler2.NewTelegramJobCrawler(storageService)

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

	app.Get("/stats", func(c fiber.Ctx) error {
		stats := storageService.GetStats()
		return c.JSON(stats)
	})

	app.Get("/stats/summary", func(c fiber.Ctx) error {
		summary, err := storageService.GetCounts()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.JSON(summary)
	})

	app.Get("/repos", func(c fiber.Ctx) error {
		includeIssues := c.Query("include_issues")
		includeCount := includeIssues == "count" || includeIssues == "true" || includeIssues == "1"

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

	app.Get("/repos/:owner/:name", func(c fiber.Ctx) error {
		owner := c.Params("owner")
		name := c.Params("name")

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

	app.Post("/crawler/start", func(c fiber.Ctx) error {
		body := c.Body()

		var req models.CrawlRequest
		if err := json.Unmarshal(body, &req); err != nil {
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

	app.Post("/crawler/telegram/jobs", func(c fiber.Ctx) error {
		type TelegramJobRequest struct {
			URLs       []string `json:"urls"`
			ClearFirst bool     `json:"clear_first,omitempty"`
			ForceParse bool     `json:"force_parse,omitempty"`
			ParseMode  string   `json:"parse_mode,omitempty"`
		}

		var req TelegramJobRequest
		if err := c.Bind().Body(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid request body"})
		}

		if len(req.URLs) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "urls is required (array of Telegram URLs)"})
		}

		var tasks []struct {
			Channel   string
			MessageID int
			IsMessage bool
			URL       string
		}

		for _, urlStr := range req.URLs {
			channel, messageID, isMessage := parseTelegramURL(urlStr)
			if channel == "" {
				log.Printf("⚠️ Invalid Telegram URL: %s", urlStr)
				continue
			}

			tasks = append(tasks, struct {
				Channel   string
				MessageID int
				IsMessage bool
				URL       string
			}{
				Channel:   channel,
				MessageID: messageID,
				IsMessage: isMessage,
				URL:       urlStr,
			})
		}

		if len(tasks) == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "no valid Telegram URLs found"})
		}

		for _, task := range tasks {
			go func(ch string, msgID int, isMsg bool, url string, clearFirst bool, force bool) {
				if isMsg {
					log.Printf("📨 Parsing specific message: %s", url)
					if err := telegramJobCrawler.CrawlTelegramMessage(ch, msgID); err != nil {
						log.Printf("❌ Error parsing message %s: %v", url, err)
					}
				} else {
					log.Printf("📢 Parsing channel: @%s (%s)", ch, url)

					if clearFirst && force {
						jobs, err := storageService.GetTelegramJobsByChannel(ch)
						if err == nil {
							deleted := 0
							for _, job := range jobs {
								key := fmt.Sprintf("telegram_job:%s:%d", job.Channel, job.MessageID)
								if err := storageService.Db.Delete(key); err == nil {
									deleted++
								}
							}
							log.Printf("🧹 Cleared %d old jobs from @%s", deleted, ch)
						}
					}

					if err := telegramJobCrawler.CrawlTelegramChannelJobs(ch); err != nil {
						log.Printf("❌ Error parsing channel @%s: %v", ch, err)
					}
				}
			}(task.Channel, task.MessageID, task.IsMessage, task.URL, req.ClearFirst, req.ForceParse)
		}

		return c.JSON(fiber.Map{
			"message":     "Telegram job crawlers started",
			"urls":        req.URLs,
			"tasks":       len(tasks),
			"force_parse": req.ForceParse,
			"clear_first": req.ClearFirst,
			"note":        "Each URL is being processed in parallel",
		})
	})

	app.Get("/telegram/jobs", func(c fiber.Ctx) error {
		page, _ := strconv.Atoi(c.Query("page", "1"))
		limit, _ := strconv.Atoi(c.Query("limit", "100"))
		offset := (page - 1) * limit

		jobs, err := storageService.GetTelegramJobsPage(limit, offset)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(jobs)
	})

	app.Get("/telegram/jobs/:channel", func(c fiber.Ctx) error {
		channel := c.Params("channel")
		query := c.Query("q", "")

		var jobs []models.TelegramJob
		var err error

		if query != "" {
			jobs, err = storageService.SearchTelegramJobs(query, channel)
		} else {
			jobs, err = storageService.GetTelegramJobsByChannel(channel)
		}

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(jobs)
	})

	app.Get("/telegram/job/:channel/:message_id", func(c fiber.Ctx) error {
		channel := c.Params("channel")
		messageIDStr := c.Params("message_id")
		messageID, err := strconv.Atoi(messageIDStr)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid message_id"})
		}

		job, err := storageService.GetTelegramJob(channel, messageID)
		if err != nil {
			return c.Status(404).JSON(fiber.Map{"error": "telegram job not found"})
		}
		return c.JSON(job)
	})

	app.Get("/telegram/jobs/search", func(c fiber.Ctx) error {
		query := c.Query("q", "")
		channel := c.Query("channel", "")

		if query == "" {
			return c.Status(400).JSON(fiber.Map{"error": "search query is required"})
		}

		jobs, err := storageService.SearchTelegramJobs(query, channel)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"query":   query,
			"channel": channel,
			"count":   len(jobs),
			"results": jobs,
		})
	})

	app.Get("/telegram/jobs/stats/:channel", func(c fiber.Ctx) error {
		channel := c.Params("channel")
		count, err := storageService.CountTelegramJobsByChannel(channel)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		jobs, err := storageService.GetTelegramJobsByChannel(channel)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		jobTypes := make(map[string]int)
		for _, job := range jobs {
			jobTypes[job.JobType]++
		}

		return c.JSON(fiber.Map{
			"channel":         channel,
			"total_jobs":      count,
			"job_types":       jobTypes,
			"latest_job_date": getLatestJobDate(jobs),
			"average_payment": getAveragePayment(jobs),
		})
	})

	app.Get("/api/routes", func(c fiber.Ctx) error {
		routes := []fiber.Map{
			{"method": "GET", "path": "/health", "description": "Health check"},
			{"method": "GET", "path": "/stats", "description": "Get database statistics"},
			{"method": "GET", "path": "/stats/summary", "description": "Get database counts summary"},
			{"method": "GET", "path": "/repos", "description": "Get all repositories (query: include_issues=count, expand=true)"},
			{"method": "GET", "path": "/repos/:owner/:name", "description": "Get specific repository"},
			{"method": "GET", "path": "/repos/:owner/:name/issues", "description": "Get repository issues"},
			{"method": "GET", "path": "/repos/:owner/:name/prs", "description": "Get repository pull requests"},
			{"method": "GET", "path": "/repos/search", "description": "Search repositories (query: language)"},
			{"method": "DELETE", "path": "/repos/:owner/:name", "description": "Delete repository"},
			{"method": "GET", "path": "/contacts", "description": "Get all contacts"},
			{"method": "GET", "path": "/contacts/:login", "description": "Get specific contact"},
			{"method": "POST", "path": "/crawler/start", "description": "Start GitHub crawler (body: start_username, max_iterations, delay_ms, github_token, use_playwright)"},
			{"method": "GET", "path": "/crawler/config", "description": "Get current crawler configuration"},
			{"method": "GET", "path": "/issues", "description": "Get all issues (query: page, limit)"},
			{"method": "POST", "path": "/crawler/telegram/jobs", "description": "Parse job posts from Telegram (body: urls=[array], clear_first, force_parse, parse_mode)"},
			{"method": "GET", "path": "/telegram/jobs", "description": "Get all Telegram jobs (query: page, limit)"},
			{"method": "GET", "path": "/telegram/jobs/:channel", "description": "Get jobs by channel (query: q for search)"},
			{"method": "GET", "path": "/telegram/job/:channel/:message_id", "description": "Get specific job post"},
			{"method": "GET", "path": "/telegram/jobs/search", "description": "Search jobs across all channels (query: q, channel)"},
			{"method": "GET", "path": "/telegram/jobs/stats/:channel", "description": "Get job statistics for channel"},
			{"method": "POST", "path": "/crawler/telegram/start", "description": "DEPRECATED: Use /crawler/telegram/jobs instead"},
			{"method": "GET", "path": "/telegram/posts", "description": "DEPRECATED: Telegram posts endpoint"},

			{"method": "GET", "path": "/api/routes", "description": "List all available endpoints"},

			// Parquet export endpoints
			{"method": "POST", "path": "/export/issues", "description": "Export all issues to Parquet format"},
			{"method": "POST", "path": "/export/pull-requests", "description": "Export all PRs to Parquet format"},
			{"method": "POST", "path": "/export/repositories", "description": "Export all repositories to Parquet format"},
			{"method": "POST", "path": "/export/all", "description": "Export all data (issues, PRs, repos) to Parquet format"},
			{"method": "POST", "path": "/export/all-jsonl", "description": "Export all data to JSONL format (better for LLM training)"},
		}
		return c.JSON(routes)
	})

	// Parquet export endpoints
	exporter := parquet.NewParquetExporter(db)

	app.Post("/export/issues", func(c fiber.Ctx) error {
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
	})

	app.Post("/export/pull-requests", func(c fiber.Ctx) error {
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
	})

	app.Post("/export/repositories", func(c fiber.Ctx) error {
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
	})

	app.Post("/export/all", func(c fiber.Ctx) error {
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
	})

	app.Post("/export/all-jsonl", func(c fiber.Ctx) error {
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
	})

	port := ":3000"
	log.Printf("Server started on %s", port)
	if err := app.Listen(port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func parseTelegramURL(urlStr string) (channel string, messageID int, isMessage bool) {
	urlStr = strings.TrimSpace(urlStr)

	if strings.Contains(urlStr, "/") && strings.Contains(urlStr, "t.me/") {
		urlStr = strings.TrimPrefix(urlStr, "https://")
		urlStr = strings.TrimPrefix(urlStr, "http://")

		if strings.HasPrefix(urlStr, "t.me/") {
			path := strings.TrimPrefix(urlStr, "t.me/")
			parts := strings.Split(path, "/")

			if len(parts) >= 1 {
				channel = parts[0]
				if len(parts) >= 2 {
					if id, err := strconv.Atoi(parts[1]); err == nil {
						messageID = id
						isMessage = true
					}
				}
			}
		}
	} else if strings.HasPrefix(urlStr, "@") || !strings.Contains(urlStr, "/") {
		channel = strings.TrimPrefix(urlStr, "@")
		isMessage = false
	}

	return channel, messageID, isMessage
}
