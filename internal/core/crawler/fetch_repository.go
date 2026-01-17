package crawler

import (
	"Fyne-on/pkg/models"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
	"sync"
	"time"
)

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
