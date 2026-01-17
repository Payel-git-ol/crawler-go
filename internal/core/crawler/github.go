package crawler

import (
	"Fyne-on/internal/core/markov"
	"net/http"
	"time"

	"Fyne-on/pkg/scraper"
	"Fyne-on/pkg/storage"
)

type GithubCrawler struct {
	storage       *storage.StorageService
	visited       map[string]bool
	client        *http.Client
	maxIterations int
	delayMs       int
	token         string
	markovChain   *markov.MarkovChain
	usePlaywright bool
	htmlScraper   *scraper.HTTPScraper
}

func NewGithubCrawler(storage *storage.StorageService) *GithubCrawler {
	return &GithubCrawler{
		storage:       storage,
		visited:       make(map[string]bool),
		client:        &http.Client{Timeout: 15 * time.Second},
		maxIterations: 20000,
		delayMs:       1000,
		markovChain:   markov.NewMarkovChain(0),
		htmlScraper:   scraper.NewHTTPScraper(15),
	}
}
