package main

import (
	"Fyne-on/pkg/database"
	"Fyne-on/pkg/models"
	"Fyne-on/pkg/storage"
	"log"
	"time"
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

	// Add test contacts
	contacts := []models.Contact{
		{
			Login:   "torvalds",
			URL:     "https://github.com/torvalds",
			Avatar:  "https://avatars.githubusercontent.com/u/1?v=4",
			Company: "Linux Foundation",
			Email:   "torvalds@linux-foundation.org",
			Bio:     "Creator of Linux",
		},
		{
			Login:   "gvanrossum",
			URL:     "https://github.com/gvanrossum",
			Avatar:  "https://avatars.githubusercontent.com/u/31?v=4",
			Company: "Microsoft",
			Email:   "",
			Bio:     "Creator of Python",
		},
	}

	for _, contact := range contacts {
		if err := storageService.SaveContact(contact); err != nil {
			log.Printf("Error saving contact %s: %v", contact.Login, err)
		} else {
			log.Printf("Saved contact: %s", contact.Login)
		}
	}

	// Add test repositories
	repos := []models.Repo{
		{
			Name:        "linux",
			Owner:       "torvalds",
			URL:         "https://github.com/torvalds/linux",
			Description: "Linux kernel source tree",
			Language:    "C",
			Stars:       50000,
		},
		{
			Name:        "cpython",
			Owner:       "python",
			URL:         "https://github.com/python/cpython",
			Description: "The Python programming language",
			Language:    "Python",
			Stars:       60000,
		},
	}

	for _, repo := range repos {
		if _, err := storageService.SaveRepo(repo); err != nil {
			log.Printf("Error saving repo %s/%s: %v", repo.Owner, repo.Name, err)
		} else {
			log.Printf("Saved repo: %s/%s", repo.Owner, repo.Name)
		}
	}

	// Add test issues
	issues := []models.Issue{
		{
			RepoID:    "torvalds/linux",
			Title:     "Fix memory leak in scheduler",
			URL:       "https://github.com/torvalds/linux/issues/1",
			State:     "open",
			Author:    "torvalds",
			Body:      "Memory leak detected in the scheduler",
			CreatedAt: time.Now(),
		},
		{
			RepoID:    "python/cpython",
			Title:     "Improve performance",
			URL:       "https://github.com/python/cpython/issues/2",
			State:     "closed",
			Author:    "gvanrossum",
			Body:      "Need to optimize interpreter performance",
			CreatedAt: time.Now().AddDate(0, 0, -1),
		},
	}

	for _, issue := range issues {
		if _, err := storageService.SaveIssue(issue); err != nil {
			log.Printf("Error saving issue %s: %v", issue.Title, err)
		} else {
			log.Printf("Saved issue: %s", issue.Title)
		}
	}

	// Add test PRs
	prs := []models.PullRequest{
		{
			RepoID:    "torvalds/linux",
			Title:     "Add new feature",
			URL:       "https://github.com/torvalds/linux/pull/100",
			State:     "open",
			Author:    "developer1",
			Body:      "Adds new feature XYZ",
			CreatedAt: time.Now(),
		},
	}

	for _, pr := range prs {
		if _, err := storageService.SavePullRequest(pr); err != nil {
			log.Printf("Error saving PR %s: %v", pr.Title, err)
		} else {
			log.Printf("Saved PR: %s", pr.Title)
		}
	}

	// Get and print stats
	stats := storageService.GetStats()
	log.Printf("Stats: %+v", stats)
}
