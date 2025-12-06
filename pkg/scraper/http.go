package scraper

// Placeholder for future Playwright implementation
// For now, we'll use HTTP client directly with GitHub API

// RepositoryIssue represents a repository issue
type RepositoryIssue struct {
	Title     string
	URL       string
	State     string // open, closed
	Author    string
	CreatedAt string
}

// Contributor represents a repository contributor
type Contributor struct {
	Login         string
	URL           string
	Avatar        string
	Contributions int
}

// HTTPScraper provides web scraping utilities
type HTTPScraper struct {
	timeout int // in seconds
}

// NewHTTPScraper creates a new HTTP scraper
func NewHTTPScraper(timeoutSec int) *HTTPScraper {
	return &HTTPScraper{
		timeout: timeoutSec,
	}
}

// Close is a no-op for HTTP scraper
func (hs *HTTPScraper) Close() error {
	return nil
}
