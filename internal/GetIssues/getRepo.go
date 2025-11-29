package GetIssues

import (
	"Fyne-on/pkg/models"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

func FetchRepo(org string, db *gorm.DB) ([]models.Repo, error) {
	var dbRepos []models.Repo
	if err := db.Find(&dbRepos).Error; err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.github.com/orgs/%s/repos", org)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
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

	var apiRepos []models.Repo
	for _, gr := range githubRepos {
		hasOpenLicense := gr.License.Key != "" && isOpenLicense(gr.License.Key)
		if !hasOpenLicense {
			apiRepos = append(apiRepos, models.Repo{
				Name:           gr.Name,
				URL:            gr.URL,
				HasOpenLicense: hasOpenLicense,
			})
		}
	}

	if !equalRepos(dbRepos, apiRepos) {
		fmt.Println("Repos changed, updating DB...")
		db.Exec("DELETE FROM repos")
		if len(apiRepos) > 0 {
			db.CreateInBatches(apiRepos, 100)
		}
		return apiRepos, nil
	}

	fmt.Println("Repos unchanged, returning from DB")
	return dbRepos, nil
}

func equalRepos(a, b []models.Repo) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name != b[i].Name || a[i].URL != b[i].URL || a[i].HasOpenLicense != b[i].HasOpenLicense {
			return false
		}
	}
	return true
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
