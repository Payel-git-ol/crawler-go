package storage

import (
	"Fyne-on/pkg/database"
	"Fyne-on/pkg/models"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type StorageService struct {
	Db *database.BadgerDB
}

func NewStorageService(db *database.BadgerDB) *StorageService {
	return &StorageService{Db: db}
}

func (s *StorageService) SaveContact(contact models.Contact) error {
	key := "contact:" + contact.Login

	if contact.Hash == "" {
		contact.Hash = database.GenerateHash(contact.Login, contact.URL)
	}
	contact.UpdatedAt = time.Now()

	return s.Db.Set(key, contact)
}

func (s *StorageService) GetContact(login string) (*models.Contact, error) {
	key := "contact:" + login
	var contact models.Contact
	err := s.Db.GetJSON(key, &contact)
	if err != nil {
		return nil, fmt.Errorf("contact not found: %w", err)
	}
	return &contact, nil
}

func (s *StorageService) SaveRepo(repo models.Repo) (bool, error) {
	key := "repo:" + repo.Owner + "/" + repo.Name

	if repo.Hash == "" {
		repo.Hash = database.GenerateHash(repo.Owner, repo.Name, repo.URL)
	}

	exists, err := s.Db.Exists(key)
	if err != nil {
		return false, err
	}

	if exists {
		var existing models.Repo
		if err := s.Db.GetJSON(key, &existing); err == nil {
			if existing.Hash == repo.Hash {
				return false, nil // No changes
			}
		}
	}

	repo.UpdatedAt = time.Now()
	return true, s.Db.Set(key, repo)
}

func (s *StorageService) GetRepo(owner, name string) (*models.Repo, error) {
	key := "repo:" + owner + "/" + name
	var repo models.Repo
	err := s.Db.GetJSON(key, &repo)
	if err != nil {
		return nil, fmt.Errorf("repo not found: %w", err)
	}
	return &repo, nil
}

func (s *StorageService) SaveIssue(issue models.Issue) (bool, error) {
	key := "issue:" + issue.RepoID + "/" + issue.ID

	if issue.Hash == "" {
		issue.Hash = database.GenerateHash(issue.RepoID, issue.ID, issue.URL)
	}

	exists, err := s.Db.Exists(key)
	if err != nil {
		return false, err
	}

	if exists {
		var existing models.Issue
		if err := s.Db.GetJSON(key, &existing); err == nil {
			if existing.Hash == issue.Hash {
				return false, nil // No changes
			}
		}
	}

	if issue.UpdatedAt.IsZero() {
		issue.UpdatedAt = time.Now()
	}

	return true, s.Db.Set(key, issue)
}

func (s *StorageService) SavePullRequest(pr models.PullRequest) (bool, error) {
	key := "pr:" + pr.RepoID + "/" + pr.ID

	if pr.Hash == "" {
		pr.Hash = database.GenerateHash(pr.RepoID, pr.ID, pr.URL)
	}

	exists, err := s.Db.Exists(key)
	if err != nil {
		return false, err
	}

	if exists {
		var existing models.PullRequest
		if err := s.Db.GetJSON(key, &existing); err == nil {
			if existing.Hash == pr.Hash {
				return false, nil // No changes
			}
		}
	}

	if pr.UpdatedAt.IsZero() {
		pr.UpdatedAt = time.Now()
	}

	return true, s.Db.Set(key, pr)
}

func (s *StorageService) GetAllRepos() ([]models.Repo, error) {
	repos := []models.Repo{}
	items, err := s.Db.GetAll("repo:")
	if err != nil {
		return nil, err
	}

	for key := range items {
		var repo models.Repo
		if err := s.Db.GetJSON(key, &repo); err == nil {
			repos = append(repos, repo)
		}
	}

	return repos, nil
}

func (s *StorageService) GetAllContacts() ([]models.Contact, error) {
	contacts := []models.Contact{}
	items, err := s.Db.GetAll("contact:")
	if err != nil {
		return nil, err
	}

	for key := range items {
		var contact models.Contact
		if err := s.Db.GetJSON(key, &contact); err == nil {
			contacts = append(contacts, contact)
		}
	}

	return contacts, nil
}

func (s *StorageService) GetRepoIssues(repoID string) ([]models.Issue, error) {
	issues := []models.Issue{}
	key := "issue:" + repoID + "/"

	err := s.Db.IterateWithPrefix(key, func(k string, v []byte) error {
		var issue models.Issue
		if err := s.Db.GetJSON(k, &issue); err == nil {
			issues = append(issues, issue)
		}
		return nil
	})

	return issues, err
}

func (s *StorageService) GetRepoPullRequests(repoID string) ([]models.PullRequest, error) {
	prs := []models.PullRequest{}
	key := "pr:" + repoID + "/"

	err := s.Db.IterateWithPrefix(key, func(k string, v []byte) error {
		var pr models.PullRequest
		if err := s.Db.GetJSON(k, &pr); err == nil {
			prs = append(prs, pr)
		}
		return nil
	})

	return prs, err
}

func (s *StorageService) GetStats() map[string]interface{} {
	repoCount := 0
	contactCount := 0
	issueCount := 0
	prCount := 0
	tgJobCount := 0

	s.Db.IterateWithPrefix("repo:", func(k string, v []byte) error {
		repoCount++
		return nil
	})

	s.Db.IterateWithPrefix("contact:", func(k string, v []byte) error {
		contactCount++
		return nil
	})

	s.Db.IterateWithPrefix("issue:", func(k string, v []byte) error {
		issueCount++
		return nil
	})

	s.Db.IterateWithPrefix("pr:", func(k string, v []byte) error {
		prCount++
		return nil
	})

	s.Db.IterateWithPrefix("telegram_job:", func(k string, v []byte) error {
		tgJobCount++
		return nil
	})

	return map[string]interface{}{
		"repositories":  repoCount,
		"contacts":      contactCount,
		"issues":        issueCount,
		"pull_requests": prCount,
		"telegram_jobs": tgJobCount,
	}
}

type StatsSummary struct {
	Contacts      int `json:"contacts"`
	Issues        int `json:"issues"`
	PullRequests  int `json:"pull_requests"`
	Repositories  int `json:"repositories"`
	TelegramPosts int `json:"telegram_posts"`
	TelegramJobs  int `json:"telegram_jobs"`
}

func (s *StorageService) GetCounts() (*StatsSummary, error) {
	contacts, err := s.Db.CountByPrefix("contact:")
	if err != nil {
		return nil, fmt.Errorf("count contacts failed: %w", err)
	}
	issues, err := s.Db.CountByPrefix("issue:")
	if err != nil {
		return nil, fmt.Errorf("count issues failed: %w", err)
	}
	prs, err := s.Db.CountByPrefix("pr:")
	if err != nil {
		return nil, fmt.Errorf("count pull_requests failed: %w", err)
	}
	repos, err := s.Db.CountByPrefix("repo:")
	if err != nil {
		return nil, fmt.Errorf("count repositories failed: %w", err)
	}
	tgPosts, err := s.Db.CountByPrefix("telegram_post:")
	if err != nil {
		return nil, fmt.Errorf("count telegram posts failed: %w", err)
	}
	tgJobs, err := s.Db.CountByPrefix("telegram_job:")
	if err != nil {
		return nil, fmt.Errorf("count telegram jobs failed: %w", err)
	}

	return &StatsSummary{
		Contacts:      contacts,
		Issues:        issues,
		PullRequests:  prs,
		Repositories:  repos,
		TelegramPosts: tgPosts,
		TelegramJobs:  tgJobs,
	}, nil
}

func (s *StorageService) DeleteRepo(owner, name string) error {
	key := "repo:" + owner + "/" + name
	repoID := owner + "/" + name

	s.Db.IterateWithPrefix("issue:"+repoID+"/", func(k string, v []byte) error {
		return s.Db.Delete(k)
	})

	s.Db.IterateWithPrefix("pr:"+repoID+"/", func(k string, v []byte) error {
		return s.Db.Delete(k)
	})

	return s.Db.Delete(key)
}

func (s *StorageService) GetIssuesPage(limit, offset int) ([]models.Issue, error) {
	const prefix = "issue:"
	out := make([]models.Issue, 0, limit)

	count := 0
	skipped := 0

	err := s.Db.IteratePrefix(prefix, func(_ []byte, v []byte) error {
		if skipped < offset {
			skipped++
			return nil
		}
		if count >= limit {
			return nil
		}

		var issue models.Issue
		if err := json.Unmarshal(v, &issue); err != nil {
			return err
		}
		out = append(out, issue)
		count++
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}
	return out, nil
}

func (s *StorageService) SaveTelegramPost(post models.TelegramPost) (bool, error) {
	key := fmt.Sprintf("telegram_post:%s:%d", post.Channel, post.MessageID)

	if post.Hash == "" {
		post.Hash = database.GenerateHash(post.Channel, fmt.Sprintf("%d", post.MessageID), post.Text)
	}

	exists, err := s.Db.Exists(key)
	if err != nil {
		return false, err
	}

	if exists {
		var existing models.TelegramPost
		if err := s.Db.GetJSON(key, &existing); err == nil {
			if existing.Hash == post.Hash {
				return false, nil // No changes
			}
		}
	}

	post.UpdatedAt = time.Now()
	if err := s.Db.Set(key, post); err != nil {
		return false, fmt.Errorf("failed to save telegram post %s: %w", key, err)
	}
	return true, nil
}

func (s *StorageService) GetTelegramPost(channel string, messageID int) (*models.TelegramPost, error) {
	key := fmt.Sprintf("telegram_post:%s:%d", channel, messageID)
	var post models.TelegramPost
	err := s.Db.GetJSON(key, &post)
	if err != nil {
		return nil, fmt.Errorf("telegram post not found: %w", err)
	}
	return &post, nil
}

func (s *StorageService) GetAllTelegramPosts() ([]models.TelegramPost, error) {
	var posts []models.TelegramPost
	err := s.Db.IteratePrefix("telegram_post:", func(key []byte, value []byte) error {
		var post models.TelegramPost
		if err := json.Unmarshal(value, &post); err != nil {
			return fmt.Errorf("failed to unmarshal telegram post: %w", err)
		}
		posts = append(posts, post)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get all telegram posts: %w", err)
	}
	return posts, nil
}

func (s *StorageService) GetTelegramPostsByChannel(channel string) ([]models.TelegramPost, error) {
	prefix := fmt.Sprintf("telegram_post:%s:", channel)
	var posts []models.TelegramPost
	err := s.Db.IteratePrefix(prefix, func(key []byte, value []byte) error {
		var post models.TelegramPost
		if err := json.Unmarshal(value, &post); err != nil {
			return fmt.Errorf("failed to unmarshal telegram post for channel %s: %w", channel, err)
		}
		posts = append(posts, post)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get telegram posts for channel %s: %w", channel, err)
	}
	return posts, nil
}

func (s *StorageService) CountTelegramPostsByChannel(channel string) (int, error) {
	prefix := fmt.Sprintf("telegram_post:%s:", channel)
	count := 0
	err := s.Db.IteratePrefix(prefix, func(_ []byte, _ []byte) error {
		count++
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count telegram posts for channel %s: %w", channel, err)
	}
	return count, nil
}

func (s *StorageService) SaveTelegramJob(job models.TelegramJob) (bool, error) {
	key := fmt.Sprintf("telegram_job:%s:%d", job.Channel, job.MessageID)

	job.Hash = database.GenerateHash(
		job.Channel,
		fmt.Sprintf("%d", job.MessageID),
		job.Title,
		job.Task,
		job.Payment,
		job.Deadline,
		time.Now().Format("20060102150405"),
	)

	job.UpdatedAt = time.Now()
	if err := s.Db.Set(key, job); err != nil {
		return false, fmt.Errorf("failed to save telegram job %s: %w", key, err)
	}

	exists, _ := s.Db.Exists(key)
	return !exists, nil
}

func (s *StorageService) GetTelegramJob(channel string, messageID int) (*models.TelegramJob, error) {
	key := fmt.Sprintf("telegram_job:%s:%d", channel, messageID)
	var job models.TelegramJob
	err := s.Db.GetJSON(key, &job)
	if err != nil {
		return nil, fmt.Errorf("telegram job not found: %w", err)
	}
	return &job, nil
}

func (s *StorageService) GetAllTelegramJobs() ([]models.TelegramJob, error) {
	var jobs []models.TelegramJob
	err := s.Db.IteratePrefix("telegram_job:", func(key []byte, value []byte) error {
		var job models.TelegramJob
		if err := json.Unmarshal(value, &job); err != nil {
			return fmt.Errorf("failed to unmarshal telegram job: %w", err)
		}
		jobs = append(jobs, job)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get all telegram jobs: %w", err)
	}
	return jobs, nil
}

func (s *StorageService) GetTelegramJobsByChannel(channel string) ([]models.TelegramJob, error) {
	prefix := fmt.Sprintf("telegram_job:%s:", channel)
	var jobs []models.TelegramJob
	err := s.Db.IteratePrefix(prefix, func(key []byte, value []byte) error {
		var job models.TelegramJob
		if err := json.Unmarshal(value, &job); err != nil {
			return fmt.Errorf("failed to unmarshal telegram job for channel %s: %w", channel, err)
		}
		jobs = append(jobs, job)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get telegram jobs for channel %s: %w", channel, err)
	}
	return jobs, nil
}

func (s *StorageService) CountTelegramJobsByChannel(channel string) (int, error) {
	prefix := fmt.Sprintf("telegram_job:%s:", channel)
	count := 0
	err := s.Db.IteratePrefix(prefix, func(_ []byte, _ []byte) error {
		count++
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count telegram jobs for channel %s: %w", channel, err)
	}
	return count, nil
}

func (s *StorageService) GetTelegramJobsPage(limit, offset int) ([]models.TelegramJob, error) {
	const prefix = "telegram_job:"
	out := make([]models.TelegramJob, 0, limit)

	count := 0
	skipped := 0

	err := s.Db.IteratePrefix(prefix, func(_ []byte, v []byte) error {
		if skipped < offset {
			skipped++
			return nil
		}
		if count >= limit {
			return nil
		}

		var job models.TelegramJob
		if err := json.Unmarshal(v, &job); err != nil {
			return err
		}
		out = append(out, job)
		count++
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list telegram jobs: %w", err)
	}
	return out, nil
}

func (s *StorageService) SearchTelegramJobs(query string, channel string) ([]models.TelegramJob, error) {
	var jobs []models.TelegramJob
	var allJobs []models.TelegramJob
	var err error

	if channel != "" {
		allJobs, err = s.GetTelegramJobsByChannel(channel)
	} else {
		allJobs, err = s.GetAllTelegramJobs()
	}

	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	for _, job := range allJobs {
		if strings.Contains(strings.ToLower(job.Title), query) ||
			strings.Contains(strings.ToLower(job.Task), query) ||
			strings.Contains(strings.ToLower(job.JobType), query) {
			jobs = append(jobs, job)
		}
	}

	return jobs, nil
}
