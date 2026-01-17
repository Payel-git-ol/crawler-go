package crawler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

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
