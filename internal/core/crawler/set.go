package crawler

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
