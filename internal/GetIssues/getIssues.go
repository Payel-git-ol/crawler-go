package GetIssues

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"Fyne-on/pkg/models"
)

func FetchIssues() ([]models.Issue, error) {
	const url = "https://api.github.com/repos/google/copybara/issues"

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

	// Организации, которые нужно исключить (коммерческие, крупные корпорации)
	excludedOrgs := map[string]bool{
		"google":     true,
		"microsoft":  true,
		"apple":      true,
		"facebook":   true,
		"amazon":     true,
		"netflix":    true,
		"twitter":    true,
		"linkedin":   true,
		"uber":       true,
		"lyft":       true,
		"airbnb":     true,
		"spotify":    true,
		"salesforce": true,
		"oracle":     true,
		"ibm":        true,
		"intel":      true,
		"nvidia":     true,
		"amd":        true,
		"qualcomm":   true,
		"cisco":      true,
		"vmware":     true,
		"redhat":     true,
		"suse":       true,
		"canonical":  true,
		"docker":     true,
		"mongodb":    true,
		"elastic":    true,
		"alexeagle":  true, // Google engineer
		"keith":      true, // Google engineer
		"tjgq":       true, // Google engineer
		"hvadehra":   true, // Google engineer
	}

	// Ключевые слова в логинах, которые указывают на коммерческие организации
	commercialKeywords := []string{
		"inc", "corp", "ltd", "llc", "co", "company", "corporation",
		"tech", "technologies", "software", "solutions", "systems",
		"cloud", "enterprise", "labs", "studio", "digital", "media",
		"group", "holdings", "ventures", "capital", "partners",
		"byte", "bits", "code", "dev", "developers", "hq", "official",
	}

	// Регулярные выражения для подозрительных логинов
	suspiciousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^[a-z]+\-[a-z]+\-[a-z]+`), // multiple-hyphens-in-name
		regexp.MustCompile(`[0-9]{4,}`),               // много цифр
		regexp.MustCompile(`^[a-z]+[0-9]{2,}$`),       // name123
	}

	var issues []models.Issue
	for _, gi := range githubIssues {
		userLogin := strings.ToLower(gi.User.Login)

		// Пропускаем если организация в черном списке
		if excludedOrgs[userLogin] {
			fmt.Printf("Filtered out: %s (excluded organization)\n", userLogin)
			continue
		}

		// Пропускаем если логин содержит коммерческие ключевые слова
		skip := false
		for _, keyword := range commercialKeywords {
			if strings.Contains(userLogin, keyword) {
				skip = true
				fmt.Printf("Filtered out: %s (commercial keyword: %s)\n", userLogin, keyword)
				break
			}
		}
		if skip {
			continue
		}

		// Пропускаем если логин соответствует подозрительным паттернам
		for _, pattern := range suspiciousPatterns {
			if pattern.MatchString(userLogin) {
				skip = true
				fmt.Printf("Filtered out: %s (suspicious pattern: %s)\n", userLogin, pattern.String())
				break
			}
		}
		if skip {
			continue
		}

		// Пропускаем если это pull request
		if strings.Contains(gi.URL, "/pull/") {
			fmt.Printf("Filtered out: %s (pull request)\n", userLogin)
			continue
		}

		state := fmt.Sprintf("#%d %s by %s", gi.Number, gi.State, gi.User.Login)
		issues = append(issues, models.Issue{
			Title: gi.Title,
			URL:   gi.URL,
			State: state,
		})
		fmt.Printf("Included: %s - %s\n", userLogin, gi.Title)
	}

	fmt.Printf("Total issues after filtering: %d\n", len(issues))

	// Перемешиваем issues
	shuffledIssues := shuffleIssues(issues)

	return shuffledIssues, nil
}

func shuffleIssues(issues []models.Issue) []models.Issue {
	rand.Seed(time.Now().UnixNano())
	shuffled := make([]models.Issue, len(issues))
	copy(shuffled, issues)

	// Алгоритм Фишера-Йейтса для перемешивания
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	fmt.Printf("Issues shuffled! First issue after shuffle: %s\n", shuffled[0].Title)
	return shuffled
}

func GetRandomIssues(limit int) ([]models.Issue, error) {
	issues, err := FetchIssues()
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > len(issues) {
		limit = len(issues)
	}

	result := issues[:limit]
	fmt.Printf("Returning %d random issues\n", len(result))
	return result, nil
}

func GetRandomIssue() (*models.Issue, error) {
	issues, err := FetchIssues()
	if err != nil {
		return nil, err
	}

	if len(issues) == 0 {
		return nil, fmt.Errorf("no issues available")
	}

	rand.Seed(time.Now().UnixNano())
	randomIndex := rand.Intn(len(issues))
	randomIssue := issues[randomIndex]

	fmt.Printf("Random issue selected: %s\n", randomIssue.Title)
	return &randomIssue, nil
}
