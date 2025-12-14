package scraper

import (
	"context"

	pw "github.com/playwright-community/playwright-go"
)

type PlaywrightScraper struct {
	ctx context.Context
}

func NewPlaywrightScraper(ctx context.Context) (*PlaywrightScraper, error) {
	if err := pw.Install(&pw.RunOptions{Verbose: false}); err != nil {
		return nil, err
	}
	return &PlaywrightScraper{ctx: ctx}, nil
}

func (ps *PlaywrightScraper) FetchTrendingDevelopers() ([]string, error) {
	out := []string{}

	pwInst, err := pw.Run()
	if err != nil {
		return out, err
	}
	defer pwInst.Stop()

	browser, err := pwInst.Chromium.Launch()
	if err != nil {
		return out, err
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		return out, err
	}

	if _, err := page.Goto("https://github.com/trending/developers"); err != nil {
		return out, err
	}

	els, err := page.QuerySelectorAll("article h1 a[href*='/']")
	if err != nil {
		return out, err
	}
	for _, el := range els {
		href, _ := el.GetAttribute("href")
		if href != "" && len(href) > 1 {
			out = append(out, href[1:])
		}
	}

	return out, nil
}
