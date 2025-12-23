package models

import (
	"encoding/json"
	"time"
)

type Contact struct {
	ID        string    `json:"id"`
	Login     string    `json:"login"`
	URL       string    `json:"url"`
	Avatar    string    `json:"avatar"`
	Company   string    `json:"company"`
	Email     string    `json:"email"`
	Location  string    `json:"location"`
	Bio       string    `json:"bio"`
	Hash      string    `json:"hash"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Repo struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Owner          string    `json:"owner"`
	URL            string    `json:"url"`
	Description    string    `json:"description"`
	Stars          int       `json:"stars"`
	Language       string    `json:"language"`
	HasOpenLicense bool      `json:"has_open_license"`
	License        string    `json:"license"`
	Hash           string    `json:"hash"`
	UpdatedAt      time.Time `json:"updated_at"`
	CreatedAt      time.Time `json:"createdAt"`
}

type Issue struct {
	ID        string    `json:"id"`
	RepoID    string    `json:"repo_id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	State     string    `json:"state"`
	Body      string    `json:"body"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Hash      string    `json:"hash"`
	Responses string    `json:"responses"`
}

type PullRequest struct {
	ID        string    `json:"id"`
	RepoID    string    `json:"repo_id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	State     string    `json:"state"`
	Body      string    `json:"body"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Hash      string    `json:"hash"`
}

type CrawlRequest struct {
	StartUsernames []string `json:"start_usernames"`
	MaxIterations  int      `json:"max_iterations"`
	DelayMs        int      `json:"delay_ms"`
	GitHubToken    string   `json:"github_token"`
	UsePlaywright  bool     `json:"use_playwright"`
}

type TelegramPost struct {
	ID        string    `json:"id"`
	Channel   string    `json:"channel"`
	MessageID int       `json:"message_id"`
	Text      string    `json:"text"`
	URL       string    `json:"url"`
	Timestamp time.Time `json:"timestamp"`
	Hash      string    `json:"hash"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TelegramJob struct {
	ID        string    `json:"id"`
	Channel   string    `json:"channel"`
	MessageID int       `json:"message_id"`
	Title     string    `json:"title"`
	JobType   string    `json:"job_type"`
	Task      string    `json:"task"`
	Payment   string    `json:"payment"`
	Deadline  string    `json:"deadline"`
	Views     string    `json:"views"`
	Date      string    `json:"date"`
	Hash      string    `json:"hash"`
	URL       string    `json:"url"`
	Timestamp time.Time `json:"timestamp"`
	UpdatedAt time.Time `json:"updated_at"`
	RawText   string    `json:"raw_text,omitempty"`
}

func (c Contact) MarshalJSON() ([]byte, error) {
	type Alias Contact
	return json.Marshal(struct {
		*Alias
	}{
		Alias: (*Alias)(&c),
	})
}

func (c *Contact) UnmarshalJSON(data []byte) error {
	type Alias Contact
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	return json.Unmarshal(data, &aux)
}
