package GetIssues

import (
	"Fyne-on/internal/ResponseIssuesService"
	"Fyne-on/pkg/models"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"net/http"
	"regexp"
	"sort"
	"strings"
)

func FetchIssues(db *gorm.DB) ([]models.Issue, error) {
	var dbIssues []models.Issue
	if err := db.Find(&dbIssues).Error; err != nil {
		return nil, err
	}

	const url = "https://api.github.com/repos/google/copybara/issues"
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

	var githubIssues []struct {
		Title  string `json:"title"`
		URL    string `json:"html_url"`
		State  string `json:"state"`
		Number int    `json:"number"`
		User   struct {
			Login string `json:"login"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&githubIssues); err != nil {
		return nil, err
	}

	excludedOrgs := map[string]bool{
		"google": true, "microsoft": true, "apple": true, "facebook": true,
		"amazon": true, "netflix": true, "twitter": true, "linkedin": true,
		"uber": true, "lyft": true, "airbnb": true, "spotify": true,
		"salesforce": true, "oracle": true, "ibm": true, "intel": true,
		"nvidia": true, "amd": true, "qualcomm": true, "cisco": true,
		"vmware": true, "redhat": true, "suse": true, "canonical": true,
		"docker": true, "mongodb": true, "elastic": true,
		"alexeagle": true, "keith": true, "tjgq": true, "hvadehra": true,
	}
	commercialKeywords := []string{"inc", "corp", "ltd", "llc", "company", "tech", "solutions"}
	suspiciousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^[a-z]+\-[a-z]+\-[a-z]+`),
		regexp.MustCompile(`[0-9]{4,}`),
		regexp.MustCompile(`^[a-z]+[0-9]{2,}$`),
	}

	var apiIssues []models.Issue
	for _, gi := range githubIssues {
		userLogin := strings.ToLower(gi.User.Login)
		if excludedOrgs[userLogin] {
			continue
		}
		skip := false
		for _, kw := range commercialKeywords {
			if strings.Contains(userLogin, kw) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		skipTitleKeywords := []string{"google", "microsoft", "apple", "facebook"}

		for _, kw := range skipTitleKeywords {
			if strings.Contains(strings.ToLower(gi.Title), kw) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		for _, pattern := range suspiciousPatterns {
			if pattern.MatchString(userLogin) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		if strings.Contains(gi.URL, "/pull/") {
			continue
		}

		state := fmt.Sprintf("#%d %s by %s", gi.Number, gi.State, gi.User.Login)
		apiIssues = append(apiIssues, models.Issue{
			Title: gi.Title,
			URL:   gi.URL,
			State: state,
		})
	}

	if !equalIssues(dbIssues, apiIssues) {
		fmt.Println("Issues changed, updating DB...")

		for _, dbIssue := range dbIssues {
			found := false
			for _, apiIssue := range apiIssues {
				if dbIssue.URL == apiIssue.URL {
					found = true
					break
				}
			}
			if !found {
				if err := ResponseIssuesService.DeleteIssue(db, dbIssue.ID); err != nil {
					fmt.Println("Ошибка удаления:", err)
				}
			}
		}

		// добавляем новые issues
		if len(apiIssues) > 0 {
			db.CreateInBatches(apiIssues, 100)
		}

		fmt.Println("Issues unchanged, returning from DB")
		var loadedIssues []models.Issue
		db.Preload("Responses").Find(&loadedIssues)
		return SortByResponses(loadedIssues), nil
	}

	fmt.Println("Issues unchanged, returning from DB")
	var loadedIssues []models.Issue
	db.Preload("Responses").Find(&loadedIssues)
	return SortByResponses(loadedIssues), nil

}

func equalIssues(a, b []models.Issue) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Title != b[i].Title || a[i].URL != b[i].URL || a[i].State != b[i].State {
			return false
		}
	}
	return true
}

func SortByResponses(issues []models.Issue) []models.Issue {
	sort.SliceStable(issues, func(i, j int) bool {
		return len(issues[i].Responses) > len(issues[j].Responses)
	})
	return issues
}
