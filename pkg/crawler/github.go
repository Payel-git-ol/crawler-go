package crawler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"Fyne-on/pkg/markov"
	"Fyne-on/pkg/models"
	"Fyne-on/pkg/scraper"
	"Fyne-on/pkg/storage"
)

type GithubCrawler struct {
	storage       *storage.StorageService
	visited       map[string]bool
	client        *http.Client
	maxIterations int
	delayMs       int
	token         string // GitHub API token (optional, for higher rate limits)

	// Added fields to fix compilation
	markovChain   *markov.MarkovChain
	usePlaywright bool

	// NEW: HTML scraper
	htmlScraper *scraper.HTTPScraper
}

// NewGithubCrawler creates a new GitHub crawler
func NewGithubCrawler(storage *storage.StorageService) *GithubCrawler {
	return &GithubCrawler{
		storage:       storage,
		visited:       make(map[string]bool),
		client:        &http.Client{Timeout: 15 * time.Second},
		maxIterations: 20000, // повышаем дефолт
		delayMs:       1000,
		markovChain:   markov.NewMarkovChain(0),
		// usePlaywright defaults to false
		htmlScraper: scraper.NewHTTPScraper(15),
	}
}

// SetGitHubToken sets GitHub API token for higher rate limits
func (gc *GithubCrawler) SetGitHubToken(token string) {
	gc.token = token
}

// SetMaxIterations sets maximum iterations
func (gc *GithubCrawler) SetMaxIterations(n int) {
	if n > 0 {
		gc.maxIterations = n
	}
}

// SetDelayMs sets delay between requests in milliseconds
func (gc *GithubCrawler) SetDelayMs(ms int) {
	if ms >= 0 {
		gc.delayMs = ms
	}
}

func (gc *GithubCrawler) UsePlaywright(v bool) {
	gc.usePlaywright = v
}

// makeRequest makes an HTTP request with proper headers
func (gc *GithubCrawler) makeRequest(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Fyne-on-Crawler/1.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	if gc.token != "" {
		req.Header.Set("Authorization", "token "+gc.token)
	}

	resp, err := gc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// FetchUserProfile fetches GitHub user profile
func (gc *GithubCrawler) FetchUserProfile(username string) (*models.Contact, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s", username)

	body, err := gc.makeRequest(url)
	if err != nil {
		return nil, err
	}

	var userData struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		HTMLURL   string `json:"html_url"`
		AvatarURL string `json:"avatar_url"`
		Company   string `json:"company"`
		Email     string `json:"email"`
		Location  string `json:"location"`
		Bio       string `json:"bio"`
	}

	if err := json.Unmarshal(body, &userData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	contact := models.Contact{
		ID:        fmt.Sprintf("%d", userData.ID),
		Login:     userData.Login,
		URL:       userData.HTMLURL,
		Avatar:    userData.AvatarURL,
		Company:   userData.Company,
		Email:     userData.Email,
		Location:  userData.Location,
		Bio:       userData.Bio,
		UpdatedAt: time.Now(),
	}

	return &contact, nil
}

// FetchUserStarredRepos fetches starred repositories
func (gc *GithubCrawler) FetchUserStarredRepos(username string) ([]models.Repo, error) {
	repos := []models.Repo{}
	page := 1

	for page <= 3 { // Limit to 3 pages (300 repos)
		url := fmt.Sprintf("https://api.github.com/users/%s/starred?per_page=100&page=%d", username, page)

		body, err := gc.makeRequest(url)
		if err != nil {
			break
		}

		var reposData []struct {
			Name  string `json:"name"`
			Owner struct {
				Login string `json:"login"`
			} `json:"owner"`
			HTMLURL         string `json:"html_url"`
			Description     string `json:"description"`
			StargazersCount int    `json:"stargazers_count"`
			Language        string `json:"language"`
			License         struct {
				Key string `json:"key"`
			} `json:"license"`
		}

		if err := json.Unmarshal(body, &reposData); err != nil {
			break
		}

		if len(reposData) == 0 {
			break
		}

		for _, rd := range reposData {
			repo := models.Repo{
				ID:          rd.Owner.Login + "/" + rd.Name,
				Name:        rd.Name,
				Owner:       rd.Owner.Login,
				URL:         rd.HTMLURL,
				Description: rd.Description,
				Stars:       rd.StargazersCount,
				Language:    rd.Language,
				License:     rd.License.Key,
				UpdatedAt:   time.Now(),
			}
			// no change here; errors logged in CrawlStart below
			repos = append(repos, repo)
		}

		page++
		time.Sleep(time.Duration(gc.delayMs) * time.Millisecond)
	}

	return repos, nil
}

// FetchRepositoryContributors fetches repository contributors
func (gc *GithubCrawler) FetchRepositoryContributors(owner, repo string) ([]models.Contact, error) {
	contacts := []models.Contact{}
	page := 1

	for page <= 2 { // Limit to 2 pages
		url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contributors?per_page=100&page=%d", owner, repo, page)

		body, err := gc.makeRequest(url)
		if err != nil {
			break
		}

		var contribData []struct {
			Login         string `json:"login"`
			ID            int    `json:"id"`
			HTMLURL       string `json:"html_url"`
			AvatarURL     string `json:"avatar_url"`
			Contributions int    `json:"contributions"`
		}

		if err := json.Unmarshal(body, &contribData); err != nil {
			break
		}

		if len(contribData) == 0 {
			break
		}

		for _, cd := range contribData {
			contact := models.Contact{
				ID:        fmt.Sprintf("%d", cd.ID),
				Login:     cd.Login,
				URL:       cd.HTMLURL,
				Avatar:    cd.AvatarURL,
				UpdatedAt: time.Now(),
			}
			contacts = append(contacts, contact)
		}

		page++
		time.Sleep(time.Duration(gc.delayMs) * time.Millisecond)
	}

	return contacts, nil
}

// FetchRepositoryIssues fetches repository issues
func (gc *GithubCrawler) FetchRepositoryIssues(owner, repo string) ([]models.Issue, error) {
	issues := []models.Issue{}
	states := []string{"open", "closed"}

	for _, state := range states {
		page := 1
		for page <= 2 { // Limit to 2 pages per state
			url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=%s&per_page=100&page=%d", owner, repo, state, page)

			body, err := gc.makeRequest(url)
			if err != nil {
				break
			}

			var issuesData []struct {
				ID      int    `json:"id"`
				Title   string `json:"title"`
				HTMLURL string `json:"html_url"`
				State   string `json:"state"`
				Body    string `json:"body"`
				User    struct {
					Login string `json:"login"`
				} `json:"user"`
				CreatedAt time.Time `json:"created_at"`
				UpdatedAt time.Time `json:"updated_at"`
			}

			if err := json.Unmarshal(body, &issuesData); err != nil {
				break
			}

			if len(issuesData) == 0 {
				break
			}

			for _, id := range issuesData {
				issue := models.Issue{
					ID:        fmt.Sprintf("%d", id.ID),
					RepoID:    owner + "/" + repo,
					Title:     id.Title,
					URL:       id.HTMLURL,
					State:     id.State,
					Body:      id.Body,
					Author:    id.User.Login,
					CreatedAt: id.CreatedAt,
					UpdatedAt: id.UpdatedAt,
				}
				issues = append(issues, issue)
			}

			page++
			time.Sleep(time.Duration(gc.delayMs) * time.Millisecond)
		}
	}

	return issues, nil
}

// FetchRepositoryPRs fetches repository pull requests
func (gc *GithubCrawler) FetchRepositoryPRs(owner, repo string) ([]models.PullRequest, error) {
	prs := []models.PullRequest{}
	states := []string{"open", "closed"}

	for _, state := range states {
		page := 1
		for page <= 2 { // Limit to 2 pages per state
			url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?state=%s&per_page=100&page=%d", owner, repo, state, page)

			body, err := gc.makeRequest(url)
			if err != nil {
				break
			}

			var prsData []struct {
				ID      int    `json:"id"`
				Title   string `json:"title"`
				HTMLURL string `json:"html_url"`
				State   string `json:"state"`
				Body    string `json:"body"`
				User    struct {
					Login string `json:"login"`
				} `json:"user"`
				CreatedAt time.Time `json:"created_at"`
				UpdatedAt time.Time `json:"updated_at"`
			}

			if err := json.Unmarshal(body, &prsData); err != nil {
				break
			}

			if len(prsData) == 0 {
				break
			}

			for _, pr := range prsData {
				pullReq := models.PullRequest{
					ID:        fmt.Sprintf("%d", pr.ID),
					RepoID:    owner + "/" + repo,
					Title:     pr.Title,
					URL:       pr.HTMLURL,
					State:     pr.State,
					Body:      pr.Body,
					Author:    pr.User.Login,
					CreatedAt: pr.CreatedAt,
					UpdatedAt: pr.UpdatedAt,
				}
				prs = append(prs, pullReq)
			}

			page++
			time.Sleep(time.Duration(gc.delayMs) * time.Millisecond)
		}
	}

	return prs, nil
}

// GetTrendingDevelopers fetches trending developers (using search API as alternative)
func (gc *GithubCrawler) GetTrendingDevelopers(language string) ([]string, error) {
	// Use search API to find popular users
	developers := []string{}

	url := fmt.Sprintf("https://api.github.com/search/users?q=followers:%%3E10&sort=followers&per_page=30")
	if language != "" {
		url = fmt.Sprintf("https://api.github.com/search/users?q=language:%s+followers:%%3E10&sort=followers&per_page=30", language)
	}

	body, err := gc.makeRequest(url)
	if err != nil {
		return nil, err
	}

	var searchResult struct {
		Items []struct {
			Login string `json:"login"`
		} `json:"items"`
	}

	if err := json.Unmarshal(body, &searchResult); err != nil {
		return nil, err
	}

	for _, item := range searchResult.Items {
		developers = append(developers, item.Login)
	}

	return developers, nil
}

// CrawlStart initiates crawling from start URL (developer username or trending)
func (gc *GithubCrawler) CrawlStart(startUsername string) error {
	visited := make(map[string]bool)
	queue := []string{startUsername}
	iteration := 0

	for len(queue) > 0 && iteration < gc.maxIterations {
		username := queue[0]
		queue = queue[1:]

		if visited[username] {
			continue
		}
		visited[username] = true
		iteration++

		fmt.Printf("Crawling: %s (iteration %d)\n", username, iteration)

		// Fetch user profile
		contact, err := gc.FetchUserProfile(username)
		if err == nil {
			gc.storage.SaveContact(*contact)
		}

		// Fetch starred repositories
		repos, err := gc.FetchUserStarredRepos(username)
		if err == nil {
			for _, repo := range repos {
				// CHANGED: capture and log errors from SaveRepo
				isNew, saveErr := gc.storage.SaveRepo(repo)
				if saveErr != nil {
					fmt.Printf("  SaveRepo failed for %s: %v\n", repo.ID, saveErr)
					continue
				}
				if isNew {
					fmt.Printf("  New repo: %s/%s\n", repo.Owner, repo.Name)

					// Fetch issues
					issues, _ := gc.FetchRepositoryIssues(repo.Owner, repo.Name)
					for _, issue := range issues {
						gc.storage.SaveIssue(issue)
					}

					// Fetch PRs
					prs, _ := gc.FetchRepositoryPRs(repo.Owner, repo.Name)
					for _, pr := range prs {
						gc.storage.SavePullRequest(pr)
					}

					// Fetch contributors and add to queue
					contributors, _ := gc.FetchRepositoryContributors(repo.Owner, repo.Name)
					for _, contrib := range contributors {
						gc.storage.SaveContact(contrib)

						// Add to queue for further crawling (limit queue to prevent explosion)
						if len(queue) < 100 && !visited[contrib.Login] {
							queue = append(queue, contrib.Login)
						}
					}
				}
			}
		}

		time.Sleep(time.Duration(gc.delayMs) * time.Millisecond)
	}

	fmt.Printf("Crawling completed. Processed %d users\n", iteration)
	return nil
}

// GetVisitedCount returns number of visited URLs
func (gc *GithubCrawler) GetVisitedCount() int {
	return len(gc.visited)
}

// CrawlStart performs HTML-based crawling starting from a username.
// It scrapes trending developers when start is empty, then user repos via HTML.
// RENAMED: avoid duplicate declaration with API-based CrawlStart
func (gc *GithubCrawler) CrawlStartHTML(startUsername string) error {
	iter := 0

	frontier := []string{}
	if startUsername != "" {
		frontier = append(frontier, startUsername)
	} else {
		trending, err := gc.htmlScraper.FetchTrendingDevelopers()
		if err == nil && len(trending) > 0 {
			frontier = append(frontier, trending...)
		} else {
			frontier = append(frontier, "torvalds")
		}
	}

	for len(frontier) > 0 && iter < gc.maxIterations {
		current := frontier[0]
		frontier = frontier[1:]

		if gc.visited[current] {
			continue
		}
		gc.visited[current] = true

		// Scrape user's repositories via HTML
		userRepos, err := gc.htmlScraper.FetchUserRepos(current)
		if err != nil {
			// continue to next user on failure
			if gc.delayMs > 0 {
				time.Sleep(time.Duration(gc.delayMs) * time.Millisecond)
			}
			continue
		}

		for _, r := range userRepos {
			repo := models.Repo{
				Name:        r.Name,
				Owner:       r.Owner,
				URL:         r.URL,
				Description: r.Description,
				Language:    r.Language,
				Stars:       r.Stars,
				// ADD: Ensure HTML-scraped repos have ID and UpdatedAt for stats
				ID: r.Owner + "/" + r.Name,
				// CHANGED: include CreatedAt; some storage/stats require it
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// CHANGED: capture and log SaveRepo errors
			isNew, saveErr := gc.storage.SaveRepo(repo)
			if saveErr != nil {
				fmt.Printf("  SaveRepo failed for %s: %v\n", repo.ID, saveErr)
				// continue to next repo if storage rejected this one
				iter++
				if gc.delayMs > 0 {
					time.Sleep(time.Duration(gc.delayMs) * time.Millisecond)
				}
				if iter >= gc.maxIterations {
					break
				}
				continue
			}
			if isNew {
				fmt.Printf("  New repo (HTML): %s\n", repo.ID)
			}

			// Basic Markov transition from user to repo
			gc.markovChain.AddTransition(current, r.Owner+"/"+r.Name)

			iter++
			if gc.delayMs > 0 {
				time.Sleep(time.Duration(gc.delayMs) * time.Millisecond)
			}
			if iter >= gc.maxIterations {
				break
			}
		}
	}

	return nil
}
