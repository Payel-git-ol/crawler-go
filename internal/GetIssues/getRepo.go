package GetIssues

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"Fyne-on/pkg/models"
)

func FetchRepo(org string) ([]models.Repo, error) {
	url := fmt.Sprintf("https://api.github.com/orgs/%s/repos", org)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "MyApp/1.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var githubRepos []struct {
		Name    string `json:"name"`
		URL     string `json:"html_url"`
		License struct {
			Key string `json:"key"`
		} `json:"license"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&githubRepos); err != nil {
		return nil, err
	}

	var repos []models.Repo
	for _, gr := range githubRepos {
		hasOpenLicense := gr.License.Key != "" && isOpenLicense(gr.License.Key)

		repos = append(repos, models.Repo{
			Name:           gr.Name,
			URL:            gr.URL,
			HasOpenLicense: hasOpenLicense,
		})
	}

	var filtered []models.Repo
	for _, r := range repos {
		if !r.HasOpenLicense {
			filtered = append(filtered, r)
		}
	}
	return filtered, nil
}

func isOpenLicense(licenseKey string) bool {
	openLicenses := []string{"mit", "apache", "gpl", "bsd", "mpl", "epl"}
	lowerKey := strings.ToLower(licenseKey)
	for _, l := range openLicenses {
		if strings.Contains(lowerKey, l) {
			return true
		}
	}
	return false
}
