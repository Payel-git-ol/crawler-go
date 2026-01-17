package main

import "github.com/gofiber/fiber/v3"

func GetRoutes(c fiber.Ctx) error {
	routes := []fiber.Map{
		{"method": "GET", "path": "/health", "description": "Health check"},
		{"method": "GET", "path": "/stats", "description": "Get database statistics"},
		{"method": "GET", "path": "/stats/summary", "description": "Get database counts summary"},
		{"method": "GET", "path": "/repos", "description": "Get all repositories (query: include_issues=count, expand=true)"},
		{"method": "GET", "path": "/repos/:owner/:name", "description": "Get specific repository"},
		{"method": "GET", "path": "/repos/:owner/:name/issues", "description": "Get repository issues"},
		{"method": "GET", "path": "/repos/:owner/:name/prs", "description": "Get repository pull requests"},
		{"method": "GET", "path": "/repos/search", "description": "Search repositories (query: language)"},
		{"method": "DELETE", "path": "/repos/:owner/:name", "description": "Delete repository"},
		{"method": "GET", "path": "/contacts", "description": "Get all contacts"},
		{"method": "GET", "path": "/contacts/:login", "description": "Get specific contact"},
		{"method": "POST", "path": "/crawler/start", "description": "Start GitHub crawler (body: start_username, max_iterations, delay_ms, github_token, use_playwright)"},
		{"method": "GET", "path": "/crawler/config", "description": "Get current crawler configuration"},
		{"method": "GET", "path": "/issues", "description": "Get all issues (query: page, limit)"},
		{"method": "POST", "path": "/crawler/telegram/jobs", "description": "Parse job posts from Telegram (body: urls=[array], clear_first, force_parse, parse_mode)"},
		{"method": "GET", "path": "/telegram/jobs", "description": "Get all Telegram jobs (query: page, limit)"},
		{"method": "GET", "path": "/telegram/jobs/:channel", "description": "Get jobs by channel (query: q for search)"},
		{"method": "GET", "path": "/telegram/job/:channel/:message_id", "description": "Get specific job post"},
		{"method": "GET", "path": "/telegram/jobs/search", "description": "Search jobs across all channels (query: q, channel)"},
		{"method": "GET", "path": "/telegram/jobs/stats/:channel", "description": "Get job statistics for channel"},
		{"method": "POST", "path": "/crawler/telegram/start", "description": "DEPRECATED: Use /crawler/telegram/jobs instead"},
		{"method": "GET", "path": "/telegram/posts", "description": "DEPRECATED: Telegram posts endpoint"},

		{"method": "GET", "path": "/api/routes", "description": "List all available endpoints"},

		{"method": "POST", "path": "/export/issues", "description": "Export all issues to Parquet format"},
		{"method": "POST", "path": "/export/pull-requests", "description": "Export all PRs to Parquet format"},
		{"method": "POST", "path": "/export/repositories", "description": "Export all repositories to Parquet format"},
		{"method": "POST", "path": "/export/all", "description": "Export all data (issues, PRs, repos) to Parquet format"},
		{"method": "POST", "path": "/export/all-jsonl", "description": "Export all data to JSONL format (better for LLM training)"},
	}
	return c.JSON(routes)
}
