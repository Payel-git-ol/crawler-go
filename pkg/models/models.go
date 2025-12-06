package models

import (
	"encoding/json"
	"time"
)

// Contact represents a GitHub user/contributor
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

// Repo represents a GitHub repository
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
}

// Issue represents a GitHub issue
type Issue struct {
	ID        string    `json:"id"`
	RepoID    string    `json:"repo_id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	State     string    `json:"state"` // open, closed
	Body      string    `json:"body"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Hash      string    `json:"hash"`
}

// PullRequest represents a GitHub PR
type PullRequest struct {
	ID        string    `json:"id"`
	RepoID    string    `json:"repo_id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	State     string    `json:"state"` // open, closed, merged
	Body      string    `json:"body"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Hash      string    `json:"hash"`
}

// MarkovState represents a state in the Markov chain traversal
type MarkovState struct {
	CurrentURL string    `json:"current_url"`
	Visited    []string  `json:"visited"`
	Timestamp  time.Time `json:"timestamp"`
}

// MarshalJSON for Contact
func (c Contact) MarshalJSON() ([]byte, error) {
	type Alias Contact
	return json.Marshal(struct {
		*Alias
	}{
		Alias: (*Alias)(&c),
	})
}

// UnmarshalJSON for Contact
func (c *Contact) UnmarshalJSON(data []byte) error {
	type Alias Contact
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(c),
	}
	return json.Unmarshal(data, &aux)
}
