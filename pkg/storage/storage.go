package storage

import (
	"Fyne-on/pkg/database"
	"Fyne-on/pkg/models"
	"encoding/json"
	"fmt"
	"time"
)

type StorageService struct {
	db *database.BadgerDB
}

// NewStorageService creates a new storage service
func NewStorageService(db *database.BadgerDB) *StorageService {
	return &StorageService{db: db}
}

// SaveContact saves or updates a contact
func (s *StorageService) SaveContact(contact models.Contact) error {
	key := "contact:" + contact.Login

	// Generate hash if not set
	if contact.Hash == "" {
		contact.Hash = database.GenerateHash(contact.Login, contact.URL)
	}
	contact.UpdatedAt = time.Now()

	return s.db.Set(key, contact)
}

// GetContact retrieves a contact
func (s *StorageService) GetContact(login string) (*models.Contact, error) {
	key := "contact:" + login
	var contact models.Contact
	err := s.db.GetJSON(key, &contact)
	if err != nil {
		return nil, fmt.Errorf("contact not found: %w", err)
	}
	return &contact, nil
}

// SaveRepo saves or updates a repository
func (s *StorageService) SaveRepo(repo models.Repo) (bool, error) {
	key := "repo:" + repo.Owner + "/" + repo.Name

	// Generate hash if not set
	if repo.Hash == "" {
		repo.Hash = database.GenerateHash(repo.Owner, repo.Name, repo.URL)
	}

	// Check if exists and hash matches
	exists, err := s.db.Exists(key)
	if err != nil {
		return false, err
	}

	if exists {
		var existing models.Repo
		if err := s.db.GetJSON(key, &existing); err == nil {
			if existing.Hash == repo.Hash {
				return false, nil // No changes
			}
		}
	}

	repo.UpdatedAt = time.Now()
	return true, s.db.Set(key, repo)
}

// GetRepo retrieves a repository
func (s *StorageService) GetRepo(owner, name string) (*models.Repo, error) {
	key := "repo:" + owner + "/" + name
	var repo models.Repo
	err := s.db.GetJSON(key, &repo)
	if err != nil {
		return nil, fmt.Errorf("repo not found: %w", err)
	}
	return &repo, nil
}

// SaveIssue saves or updates an issue
func (s *StorageService) SaveIssue(issue models.Issue) (bool, error) {
	key := "issue:" + issue.RepoID + "/" + issue.ID

	// Generate hash if not set
	if issue.Hash == "" {
		issue.Hash = database.GenerateHash(issue.RepoID, issue.ID, issue.URL)
	}

	// Check if exists and hash matches
	exists, err := s.db.Exists(key)
	if err != nil {
		return false, err
	}

	if exists {
		var existing models.Issue
		if err := s.db.GetJSON(key, &existing); err == nil {
			if existing.Hash == issue.Hash {
				return false, nil // No changes
			}
		}
	}

	if issue.UpdatedAt.IsZero() {
		issue.UpdatedAt = time.Now()
	}

	return true, s.db.Set(key, issue)
}

// SavePullRequest saves or updates a pull request
func (s *StorageService) SavePullRequest(pr models.PullRequest) (bool, error) {
	key := "pr:" + pr.RepoID + "/" + pr.ID

	// Generate hash if not set
	if pr.Hash == "" {
		pr.Hash = database.GenerateHash(pr.RepoID, pr.ID, pr.URL)
	}

	// Check if exists and hash matches
	exists, err := s.db.Exists(key)
	if err != nil {
		return false, err
	}

	if exists {
		var existing models.PullRequest
		if err := s.db.GetJSON(key, &existing); err == nil {
			if existing.Hash == pr.Hash {
				return false, nil // No changes
			}
		}
	}

	if pr.UpdatedAt.IsZero() {
		pr.UpdatedAt = time.Now()
	}

	return true, s.db.Set(key, pr)
}

// GetAllRepos retrieves all repositories
func (s *StorageService) GetAllRepos() ([]models.Repo, error) {
	repos := []models.Repo{}
	items, err := s.db.GetAll("repo:")
	if err != nil {
		return nil, err
	}

	for key := range items {
		var repo models.Repo
		if err := s.db.GetJSON(key, &repo); err == nil {
			repos = append(repos, repo)
		}
	}

	return repos, nil
}

// GetAllContacts retrieves all contacts
func (s *StorageService) GetAllContacts() ([]models.Contact, error) {
	contacts := []models.Contact{}
	items, err := s.db.GetAll("contact:")
	if err != nil {
		return nil, err
	}

	for key := range items {
		var contact models.Contact
		if err := s.db.GetJSON(key, &contact); err == nil {
			contacts = append(contacts, contact)
		}
	}

	return contacts, nil
}

// GetRepoIssues retrieves all issues for a repository
func (s *StorageService) GetRepoIssues(repoID string) ([]models.Issue, error) {
	issues := []models.Issue{}
	key := "issue:" + repoID + "/"

	err := s.db.IterateWithPrefix(key, func(k string, v []byte) error {
		var issue models.Issue
		if err := s.db.GetJSON(k, &issue); err == nil {
			issues = append(issues, issue)
		}
		return nil
	})

	return issues, err
}

// GetRepoPullRequests retrieves all pull requests for a repository
func (s *StorageService) GetRepoPullRequests(repoID string) ([]models.PullRequest, error) {
	prs := []models.PullRequest{}
	key := "pr:" + repoID + "/"

	err := s.db.IterateWithPrefix(key, func(k string, v []byte) error {
		var pr models.PullRequest
		if err := s.db.GetJSON(k, &pr); err == nil {
			prs = append(prs, pr)
		}
		return nil
	})

	return prs, err
}

// GetStats returns database statistics
func (s *StorageService) GetStats() map[string]interface{} {
	repoCount := 0
	contactCount := 0
	issueCount := 0
	prCount := 0

	s.db.IterateWithPrefix("repo:", func(k string, v []byte) error {
		repoCount++
		return nil
	})

	s.db.IterateWithPrefix("contact:", func(k string, v []byte) error {
		contactCount++
		return nil
	})

	s.db.IterateWithPrefix("issue:", func(k string, v []byte) error {
		issueCount++
		return nil
	})

	s.db.IterateWithPrefix("pr:", func(k string, v []byte) error {
		prCount++
		return nil
	})

	return map[string]interface{}{
		"repositories":  repoCount,
		"contacts":      contactCount,
		"issues":        issueCount,
		"pull_requests": prCount,
	}
}

// DeleteRepo deletes a repository and its related data
func (s *StorageService) DeleteRepo(owner, name string) error {
	key := "repo:" + owner + "/" + name
	repoID := owner + "/" + name

	// Delete issues
	s.db.IterateWithPrefix("issue:"+repoID+"/", func(k string, v []byte) error {
		return s.db.Delete(k)
	})

	// Delete PRs
	s.db.IterateWithPrefix("pr:"+repoID+"/", func(k string, v []byte) error {
		return s.db.Delete(k)
	})

	// Delete repo
	return s.db.Delete(key)
}

// StatsSummary is a compact JSON with aggregated counts.
// It matches the exact format you requested.
type StatsSummary struct {
	Contacts     int `json:"contacts"`
	Issues       int `json:"issues"`
	PullRequests int `json:"pull_requests"`
	Repositories int `json:"repositories"`
}

// GetCounts returns a StatsSummary by counting key prefixes in the KV store.
func (s *StorageService) GetCounts() (*StatsSummary, error) {
	contacts, err := s.db.CountByPrefix("contact:")
	if err != nil {
		return nil, fmt.Errorf("count contacts failed: %w", err)
	}
	issues, err := s.db.CountByPrefix("issue:")
	if err != nil {
		return nil, fmt.Errorf("count issues failed: %w", err)
	}
	prs, err := s.db.CountByPrefix("pr:")
	if err != nil {
		return nil, fmt.Errorf("count pull_requests failed: %w", err)
	}
	repos, err := s.db.CountByPrefix("repo:")
	if err != nil {
		return nil, fmt.Errorf("count repositories failed: %w", err)
	}

	return &StatsSummary{
		Contacts:     contacts,
		Issues:       issues,
		PullRequests: prs,
		Repositories: repos,
	}, nil
}

// INSERT: return all issues across all repositories
func (s *StorageService) GetAllIssues() ([]models.Issue, error) {
	const prefix = "issue:"
	out := make([]models.Issue, 0, 256)

	err := s.db.IteratePrefix(prefix, func(_ []byte, v []byte) error {
		var issue models.Issue
		if err := json.Unmarshal(v, &issue); err != nil {
			return err
		}
		out = append(out, issue)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}
	return out, nil
}
