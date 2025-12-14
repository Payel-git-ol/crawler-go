package crawler

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
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

func (gc *GithubCrawler) SetGitHubToken(token string) {
	gc.token = token
}

func (gc *GithubCrawler) SetMaxIterations(n int) {
	if n > 0 {
		gc.maxIterations = n
	}
}

func (gc *GithubCrawler) SetDelayMs(ms int) {
	if ms >= 0 {
		gc.delayMs = ms
	}
}

func (gc *GithubCrawler) UsePlaywright(v bool) {
	gc.usePlaywright = v
}

func (gc *GithubCrawler) makeRequest(url string) ([]byte, error) {
	maxRetries := 5
	retryDelay := time.Second * 5

	for i := 0; i < maxRetries; i++ {
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

		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
			remaining := resp.Header.Get("X-RateLimit-Remaining")
			resetTime := resp.Header.Get("X-RateLimit-Reset")

			if remaining == "0" && resetTime != "" {
				resetUnix, _ := strconv.ParseInt(resetTime, 10, 64)
				sleepDuration := time.Until(time.Unix(resetUnix, 0)) + time.Second

				log.Printf("Rate limit hit. Sleeping for %v...", sleepDuration)
				resp.Body.Close()
				time.Sleep(sleepDuration)
				continue
			}

			log.Printf("Abuse detection mechanism triggered. Retrying in %v...", retryDelay)
			resp.Body.Close()
			time.Sleep(retryDelay)
			retryDelay *= 2
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("status code: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		return body, nil
	}

	return nil, fmt.Errorf("max retries exceeded for url: %s", url)
}

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

func (gc *GithubCrawler) FetchUserStarredRepos(username string) ([]models.Repo, error) {
	repos := []models.Repo{}
	page := 1

	for page <= 3 {
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

			repos = append(repos, repo)
		}

		page++
		time.Sleep(time.Duration(gc.delayMs) * time.Millisecond)
	}

	return repos, nil
}

func (gc *GithubCrawler) FetchRepositoryContributors(owner, repo string) ([]models.Contact, error) {
	contacts := []models.Contact{}
	page := 1

	for page <= 2 {
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

func (gc *GithubCrawler) FetchRepositoryIssues(owner, repo string, saveFunc func(models.Issue) error) error {
	states := []string{"open", "closed"}

	for _, state := range states {
		for page := 1; ; page++ {
			url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues?state=%s&per_page=100&page=%d", owner, repo, state, page)
			log.Printf("  Fetching %s issues page %d for %s/%s", state, page, owner, repo)

			body, err := gc.makeRequest(url)

			if err != nil {
				log.Printf("Error fetching issues page %d for %s/%s (state: %s): %v", page, owner, repo, state, err)
				return err
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
				PullReq   *struct{} `json:"pull_request,omitempty"`
			}

			if err := json.Unmarshal(body, &issuesData); err != nil {
				log.Printf("Error unmarshaling issues data for %s/%s: %v", owner, repo, err)
				return err
			}

			if len(issuesData) == 0 {
				break
			}

			count := 0
			for _, id := range issuesData {
				if id.PullReq != nil {
					continue
				}

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

				if err := saveFunc(issue); err != nil {
					log.Printf("Failed to save issue %s: %v", issue.ID, err)
				}

				if err := json.Unmarshal(body, &issuesData); err != nil {
					log.Printf("Error unmarshaling issues data for %s/%s: %v\nBody: %s", owner, repo, err, string(body))
					return err
				}

				if len(issuesData) == 0 {
					log.Printf("No issues returned for %s/%s (state: %s, page: %d). Body: %s", owner, repo, state, page, string(body))
					break
				}

				count++
			}

			log.Printf("  Saved %d issues from page %d (state: %s)", count, page, state)

			if gc.delayMs > 0 {
				time.Sleep(time.Duration(gc.delayMs) * time.Millisecond)
			}
		}
	}

	return nil
}

func (gc *GithubCrawler) FetchRepositoryPRs(owner, repo string) ([]models.PullRequest, error) {
	prs := []models.PullRequest{}
	states := []string{"open", "closed"}

	for _, state := range states {
		for page := 1; ; page++ {
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

			time.Sleep(time.Duration(gc.delayMs) * time.Millisecond)
		}
	}

	return prs, nil
}

func (gc *GithubCrawler) FetchUserRepos(username string) ([]models.Repo, error) {
	repos := []models.Repo{}
	page := 1
	for {
		url := fmt.Sprintf("https://api.github.com/users/%s/repos?per_page=100&page=%d", username, page)
		body, err := gc.makeRequest(url)
		if err != nil {
			break
		}
		var reposData []struct {
			Name  string `json:"name"`
			Owner struct {
				Login string `json:"login"`
			} `json:"owner"`
			HTMLURL     string `json:"html_url"`
			Description string `json:"description"`
			Stars       int    `json:"stargazers_count"`
			Language    string `json:"language"`
			License     struct {
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
				Stars:       rd.Stars,
				Language:    rd.Language,
				License:     rd.License.Key,
				UpdatedAt:   time.Now(),
			}
			repos = append(repos, repo)
		}
		page++
		time.Sleep(time.Duration(gc.delayMs) * time.Millisecond)
	}
	return repos, nil
}

func (gc *GithubCrawler) FetchOrgReposHTML(org string) ([]models.Repo, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	repos := []models.Repo{}
	seen := make(map[string]bool)

	for page := 1; page <= 50; page++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			url := fmt.Sprintf("https://github.com/orgs/%s/repositories?page=%d", org, p)
			doc, err := gc.htmlScraper.FetchDocument(url)
			if err != nil {
				log.Printf("HTML fetch failed for %s page %d: %v", org, p, err)
				return
			}

			found := 0

			doc.Find("a[data-hovercard-type='repository']").Each(func(i int, s *goquery.Selection) {
				href, _ := s.Attr("href")
				parts := strings.Split(href, "/")
				if len(parts) < 3 {
					return
				}
				id := parts[1] + "/" + parts[2]

				mu.Lock()
				if seen[id] {
					mu.Unlock()
					return
				}
				seen[id] = true
				repo := models.Repo{
					Owner:     parts[1],
					Name:      parts[2],
					URL:       "https://github.com" + href,
					ID:        id,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				repos = append(repos, repo)
				gc.storage.SaveRepo(repo)
				mu.Unlock()
				found++
			})

			doc.Find("h3 a").Each(func(i int, s *goquery.Selection) {
				href, _ := s.Attr("href")
				parts := strings.Split(href, "/")
				if len(parts) < 3 {
					return
				}
				id := parts[1] + "/" + parts[2]

				mu.Lock()
				if seen[id] {
					mu.Unlock()
					return
				}
				seen[id] = true
				repo := models.Repo{
					Owner:     parts[1],
					Name:      parts[2],
					URL:       "https://github.com" + href,
					ID:        id,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				repos = append(repos, repo)
				gc.storage.SaveRepo(repo)
				mu.Unlock()
				found++
			})

			doc.Find("li.Box-row a").Each(func(i int, s *goquery.Selection) {
				href, _ := s.Attr("href")
				if !strings.Contains(href, "/"+org+"/") {
					return
				}
				parts := strings.Split(href, "/")
				if len(parts) < 3 {
					return
				}
				id := parts[1] + "/" + parts[2]

				mu.Lock()
				if seen[id] {
					mu.Unlock()
					return
				}
				seen[id] = true
				repo := models.Repo{
					Owner:     parts[1],
					Name:      parts[2],
					URL:       "https://github.com" + href,
					ID:        id,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				repos = append(repos, repo)
				gc.storage.SaveRepo(repo)
				mu.Unlock()
				found++
			})

			log.Printf("Page %d for %s: found %d repos", p, org, found)
		}(page)
	}

	wg.Wait()
	return repos, nil
}

func parseStars(s string) int {
	s = strings.ReplaceAll(s, ",", "")
	val, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return val
}

func (gc *GithubCrawler) GetTrendingDevelopers(language string) ([]string, error) {
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

		log.Printf("Crawling: %s (iteration %d)\n", username, iteration)

		contact, err := gc.FetchUserProfile(username)
		if err == nil {
			gc.storage.SaveContact(*contact)
		}

		repos, err := gc.FetchUserRepos(username)
		if err == nil {
			for _, repo := range repos {
				_, saveErr := gc.storage.SaveRepo(repo)
				if saveErr != nil {
					log.Printf("  SaveRepo failed for %s: %v\n", repo.ID, saveErr)
					continue
				}

				repoID := repo.Owner + "/" + repo.Name
				log.Printf("  Processing repo: %s\n", repoID)

				issueErr := gc.FetchRepositoryIssues(repo.Owner, repo.Name, func(issue models.Issue) error {
					_, err := gc.storage.SaveIssue(issue)
					return err
				})

				if issueErr != nil {
					log.Printf("  Error processing issues for %s: %v", repoID, issueErr)
				}

				prs, _ := gc.FetchRepositoryPRs(repo.Owner, repo.Name)
				for _, pr := range prs {
					gc.storage.SavePullRequest(pr)
				}

				contributors, _ := gc.FetchRepositoryContributors(repo.Owner, repo.Name)
				for _, contrib := range contributors {
					gc.storage.SaveContact(contrib)

					if len(queue) < 100 && !visited[contrib.Login] {
						queue = append(queue, contrib.Login)
					}
				}
			}
		}

		time.Sleep(time.Duration(gc.delayMs) * time.Millisecond)
	}

	log.Printf("Crawling completed. Processed %d users\n", iteration)
	return nil
}

func (gc *GithubCrawler) CrawlStartOrgsHTML(orgs []string) error {
	iter := 0

	for _, org := range orgs {
		log.Printf("Crawling org: %s", org)

		repos, err := gc.FetchOrgReposHTML(org)
		if err != nil {
			log.Printf("Failed to fetch repos for %s: %v", org, err)
			continue
		}

		for _, repo := range repos {
			isNew, saveErr := gc.storage.SaveRepo(repo)
			if saveErr != nil {
				log.Printf("SaveRepo failed for %s: %v", repo.ID, saveErr)
				continue
			}
			if isNew {
				log.Printf("New repo saved: %s", repo.ID)
			}

			iter++
			if iter >= gc.maxIterations {
				log.Printf("Reached max iterations (%d)", gc.maxIterations)
				return nil
			}
		}

		time.Sleep(time.Duration(gc.delayMs) * time.Millisecond)
	}

	log.Printf("HTML crawling completed. Saved %d repos", iter)
	return nil
}

func (gc *GithubCrawler) GetVisitedCount() int {
	return len(gc.visited)
}

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

		userRepos, err := gc.htmlScraper.FetchUserRepos(current)
		if err != nil {
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
				ID:          r.Owner + "/" + r.Name,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			isNew, saveErr := gc.storage.SaveRepo(repo)
			if saveErr != nil {
				fmt.Printf("  SaveRepo failed for %s: %v\n", repo.ID, saveErr)
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
