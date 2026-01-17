package crawler

import (
	"Fyne-on/pkg/models"
	"encoding/json"
	"fmt"
	"time"
)

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
