package main

import (
	"Fyne-on/pkg/models"
	"crypto/sha256"
	"encoding/hex"
	"github.com/gofiber/fiber/v3"
	"log"
	"sync"
)

func GetRepos(c fiber.Ctx) error {
	includeIssues := c.Query("include_issues")
	includeCount := includeIssues == "count" || includeIssues == "true" || includeIssues == "1"

	expandQ := c.Query("expand")
	expand := expandQ == "1" || expandQ == "true" || expandQ == "full"

	repos, err := StorageService.GetAllRepos()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	result := make([]fiber.Map, 0, len(repos))

	var wg sync.WaitGroup

	for _, repo := range repos {

		wg.Add(1)

		go func(repo models.Repo) {
			defer wg.Done()

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
				issues, err := StorageService.GetRepoIssues(repo.Owner + "/" + repo.Name)
				if err == nil {
					item["issues_count"] = len(issues)
				} else {
					item["issues_count"] = 0
					log.Printf("failed to load issues for %s/%s: %v", repo.Owner, repo.Name, err)
				}
			}

			result = append(result, item)
		}(repo)

	}

	wg.Wait()
	return c.JSON(result)
}

func GetReposOwnerName(c fiber.Ctx) error {
	owner := c.Params("owner")
	name := c.Params("name")

	expandQ := c.Query("expand")
	expand := expandQ == "1" || expandQ == "true" || expandQ == "full"

	repo, err := StorageService.GetRepo(owner, name)
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
}

func DeleteReposOwnerName(c fiber.Ctx) error {
	owner := c.Params("owner")
	name := c.Params("name")

	if err := StorageService.DeleteRepo(owner, name); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "repository deleted"})
}

func GetReposOwnerNameIssues(c fiber.Ctx) error {
	owner := c.Params("owner")
	name := c.Params("name")
	repoID := owner + "/" + name

	issues, err := StorageService.GetRepoIssues(repoID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(issues)
}

func GetReposOwnerNamePrs(c fiber.Ctx) error {
	owner := c.Params("owner")
	name := c.Params("name")
	repoID := owner + "/" + name

	prs, err := StorageService.GetRepoPullRequests(repoID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(prs)
}

func ReposSearch(c fiber.Ctx) error {
	language := c.Query("language")

	repos, err := StorageService.GetAllRepos()
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
}
