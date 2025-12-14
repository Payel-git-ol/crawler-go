package scraper

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// HTTPScraper provides web scraping utilities
type HTTPScraper struct {
	client *http.Client
}

func NewHTTPScraper(timeoutSec int) *HTTPScraper {
	return &HTTPScraper{
		client: &http.Client{Timeout: time.Duration(timeoutSec) * time.Second},
	}
}

// FetchDocument загружает страницу и возвращает goquery.Document
func (s *HTTPScraper) FetchDocument(url string) (*goquery.Document, error) {
	resp, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

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

// FetchTrendingDevelopers scrapes GitHub Trending Developers (HTML)
func (hs *HTTPScraper) FetchTrendingDevelopers() ([]string, error) {
	resp, err := hs.client.Get("https://github.com/trending/developers")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	out := []string{}
	doc.Find("article h1 a[href]").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok && strings.HasPrefix(href, "/") {
			login := strings.TrimPrefix(href, "/")
			if login != "" {
				out = append(out, login)
			}
		}
	})
	return unique(out), nil
}

// FetchUserRepos scrapes a user's repositories page (HTML)
func (hs *HTTPScraper) FetchUserRepos(username string) ([]struct {
	Name        string
	Owner       string
	URL         string
	Description string
	Language    string
	Stars       int
}, error) {
	url := fmt.Sprintf("https://github.com/%s?tab=repositories", username)
	resp, err := hs.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	repos := []struct {
		Name        string
		Owner       string
		URL         string
		Description string
		Language    string
		Stars       int
	}{}

	// NOTE: GitHub markup can change; selectors are best-effort.
	// Try anchors inside repo cards
	doc.Find("h3 a[href*='/'']").Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok || !strings.HasPrefix(href, "/") {
			return
		}
		parts := strings.Split(strings.TrimPrefix(href, "/"), "/")
		if len(parts) < 2 {
			return
		}
		owner := parts[0]
		name := parts[1]
		fullURL := "https://github.com/" + owner + "/" + name

		desc := strings.TrimSpace(s.Closest("article").Find("p").First().Text())
		lang := strings.TrimSpace(s.Closest("article").Find("[itemprop='programmingLanguage']").First().Text())
		// Stars often not present on list; default to 0
		repos = append(repos, struct {
			Name        string
			Owner       string
			URL         string
			Description string
			Language    string
			Stars       int
		}{
			Name:        name,
			Owner:       owner,
			URL:         fullURL,
			Description: desc,
			Language:    lang,
			Stars:       0,
		})
	})

	return repos, nil
}

// Close is a no-op for HTTP scraper
func (hs *HTTPScraper) Close() error {
	return nil
}

func unique(list []string) []string {
	seen := make(map[string]struct{}, len(list))
	out := make([]string, 0, len(list))
	for _, v := range list {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
