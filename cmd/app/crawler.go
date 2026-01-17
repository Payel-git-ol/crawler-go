package main

import (
	"Fyne-on/internal/core/crawler"
	"Fyne-on/pkg/models"
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	"log"
)

type Config struct {
	StartUsername string
	MaxIterations int
	DelayMs       int
	TokenSet      bool
	UsePlaywright bool
}

var githubCrawler *crawler.GithubCrawler

func CrawlerStart(c fiber.Ctx) error {
	body := c.Body()

	currentCrawlerConfig := Config{
		StartUsername: "",
		MaxIterations: 20000,
		DelayMs:       1000,
		TokenSet:      false,
		UsePlaywright: false,
	}

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
}

func CrawlerConfig(c fiber.Ctx) error {
	currentCrawlerConfig := Config{
		StartUsername: "",
		MaxIterations: 20000,
		DelayMs:       1000,
		TokenSet:      false,
		UsePlaywright: false,
	}

	return c.JSON(fiber.Map{
		"start_username": currentCrawlerConfig.StartUsername,
		"max_iterations": currentCrawlerConfig.MaxIterations,
		"delay_ms":       currentCrawlerConfig.DelayMs,
		"token_set":      currentCrawlerConfig.TokenSet,
		"use_playwright": currentCrawlerConfig.UsePlaywright,
	})
}
