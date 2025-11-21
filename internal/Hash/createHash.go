package Hash

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func HasOpenLicense(repoURL string) bool {
	parts := strings.Split(repoURL, "/")
	if len(parts) < 5 {
		return false
	}
	owner := parts[3]
	repo := parts[4]

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}

	req.Header.Set("User-Agent", "MyApp/1.0")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	defer resp.Body.Close()

	var repoData struct {
		License struct {
			Key string `json:"key"`
		} `json:"license"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&repoData); err != nil {
		return false
	}

	return isOpenLicense(repoData.License.Key)
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
