package crawler

import (
	"Fyne-on/pkg/models"
	"log"
	"sync"
	"time"
)

func (gc *GithubCrawler) CrawlStart(startUsername string) error {
	visited := make(map[string]bool)
	queue := []string{startUsername}
	iteration := 0

	userChan := make(chan string, 10)
	done := make(chan bool)

	var mu sync.Mutex

	for i := 0; i < 3; i++ {
		go func(workerID int) {
			for username := range userChan {
				mu.Lock()
				if visited[username] {
					mu.Unlock()
					continue
				}
				visited[username] = true
				iteration++
				currentIter := iteration
				mu.Unlock()

				if currentIter > gc.maxIterations {
					done <- true
					return
				}

				log.Printf("Worker %d: %s (iter %d)", workerID, username, currentIter)
				gc.processUserSimple(username, &queue, &mu)

				if gc.delayMs > 0 {
					time.Sleep(time.Duration(gc.delayMs) * time.Millisecond)
				}
			}
		}(i)
	}

	go func() {
		for iteration < gc.maxIterations && len(queue) > 0 {
			mu.Lock()
			username := queue[0]
			queue = queue[1:]
			mu.Unlock()

			userChan <- username
		}
		close(userChan)
		done <- true
	}()

	<-done
	log.Printf("Crawling completed. Processed %d users", iteration)
	return nil
}

func (gc *GithubCrawler) processUserSimple(username string, queue *[]string, mu *sync.Mutex) {
	contact, err := gc.FetchUserProfile(username)
	if err == nil {
		gc.storage.SaveContact(*contact)
	}

	repos, err := gc.FetchUserRepos(username)
	if err != nil {
		return
	}

	for _, repo := range repos {
		_, saveErr := gc.storage.SaveRepo(repo)
		if saveErr != nil {
			continue
		}

		go func(r models.Repo) {
			// Issues
			gc.FetchRepositoryIssues(r.Owner, r.Name, func(issue models.Issue) error {
				gc.storage.SaveIssue(issue)
				return nil
			})

			// PRs
			prs, _ := gc.FetchRepositoryPRs(r.Owner, r.Name)
			for _, pr := range prs {
				gc.storage.SavePullRequest(pr)
			}

			// Contributors
			contributors, _ := gc.FetchRepositoryContributors(r.Owner, r.Name)
			for _, contrib := range contributors {
				gc.storage.SaveContact(contrib)

				mu.Lock()
				*queue = append(*queue, contrib.Login)
				mu.Unlock()
			}
		}(repo)
	}
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
