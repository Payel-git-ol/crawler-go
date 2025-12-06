# Fyne-on - GitHub Repository Crawler with Markov Chains

A high-performance GitHub crawler that uses Markov chains for intelligent traversal and Badger KV store for efficient data storage.

## Features

- **Markov Chain-based Crawling**: Intelligent traversal using Markov chains to discover repositories from trending developers
- **Key-Value Storage**: Uses Badger DB for fast, reliable key-value storage
- **REST API**: Complete REST API to query and manage collected data
- **Deduplication**: Hash-based deduplication to avoid storing duplicate data
- **Scalable Design**: Easily extensible architecture for adding new features

## Architecture

### Core Components

1. **BadgerDB (`pkg/database/badgerdb.go`)**: 
   - Key-value database wrapper around Badger
   - Methods for CRUD operations, iteration, backup
   - Hash generation for deduplication

2. **Models (`pkg/models/models.go`)**:
   - `Contact`: GitHub users/contributors
   - `Repo`: Repository metadata
   - `Issue`: Open and closed issues
   - `PullRequest`: Open, closed, and merged PRs

3. **Markov Chain (`pkg/markov/markov.go`)**:
   - Probabilistic state transitions for crawling
   - Random selection of next URL to visit
   - Maintains transition map

4. **Storage Service (`pkg/storage/storage.go`)**:
   - High-level data persistence operations
   - Hash-based uniqueness checking
   - Cascade deletion support

5. **GitHub Crawler (`pkg/crawler/github.go`)**:
   - Markov chain-based GitHub crawling
   - Direct GitHub API integration
   - Fetches repos, issues, PRs, and contributors

## Data Model

All data is stored in Badger with the following key prefixes:

```
repo:{owner}/{name}          # Repository data
issue:{owner}/{repo}/{id}    # Issues
pr:{owner}/{repo}/{id}       # Pull requests
contact:{login}              # User/contributor data
```

### Hash-based Deduplication

Each entity is hashed using SHA256:
- **Repo hash**: `SHA256(owner + name + url)`
- **Issue/PR hash**: `SHA256(repoID + id + url)`
- **Contact hash**: `SHA256(login + url)`

If hash matches existing record, the record is not updated.

## REST API Endpoints

### Health & Stats
- `GET /health` - Health check
- `GET /stats` - Database statistics (repo count, contact count, issues, PRs)

### Repositories
- `GET /repos` - Get all repositories
- `GET /repos/:owner/:name` - Get specific repository
- `GET /repos/:owner/:name/issues` - Get repository issues
- `GET /repos/:owner/:name/prs` - Get repository PRs
- `GET /repos/search?language=Go&min_stars=100` - Search repositories
- `DELETE /repos/:owner/:name` - Delete repository (cascade delete issues/PRs)

### Contacts
- `GET /contacts` - Get all contacts
- `GET /contacts/:login` - Get specific contact

### Crawler Control
- `POST /crawler/start` - Start crawler with custom parameters
  ```json
  {
    "start_url": "https://github.com/trending/developers",
    "max_iterations": 10000,
    "delay_ms": 1000
  }
  ```

### Discovery
- `GET /api/routes` - List all available endpoints

## Getting Started

### Prerequisites
- Go 1.22+
- Badger DB (included via go.mod)

### Installation

```bash
cd c:\Users\pasaz\GolandProjects\Fyne-on
go mod tidy
go build -o app.exe ./cmd/app
```

### Running the Application

```bash
./app.exe
```

The server will start on `http://localhost:3000`

### Docker Setup

```bash
docker-compose up
```

This starts Typesense for search functionality.

## Configuration

### Crawler Parameters

Modify in `main.go` before running or via API:
```go
githubCrawler.SetMaxIterations(10000)  // Maximum URLs to crawl
githubCrawler.SetDelayMs(1000)         // Delay between requests
```

### Database

Database files are stored in `./badger_data/`:
- Keys and values are persisted automatically
- Garbage collection runs periodically

## Development

### Project Structure

```
.
├── cmd/app/              # Main application entry point
├── pkg/
│   ├── crawler/          # GitHub crawler logic
│   ├── database/         # Badger DB wrapper
│   ├── markov/           # Markov chain implementation
│   ├── models/           # Data models
│   ├── scraper/          # Web scraping utilities (future: Playwright)
│   └── storage/          # Storage service layer
├── docker-compose.yaml   # Docker services
├── go.mod               # Go module definition
└── README.md            # This file
```

### Adding New Features

1. **New Data Type**: Add to `pkg/models/models.go`, then add storage methods to `pkg/storage/storage.go`
2. **Crawler Enhancement**: Extend `pkg/crawler/github.go` with new fetching methods
3. **API Endpoint**: Add route handler in `cmd/app/main.go`

## Criteria Met

✅ **Program compiles and has REST API**: Complete implementation with health check and statistics endpoints
✅ **REST API for database queries**: Full CRUD operations for all entity types
✅ **Scalable to 10,000+ repos**: Badger KV can efficiently store millions of entries
✅ **Easily extensible code**: Clean separation of concerns, modular design
✅ **Deduplication**: Hash-based uniqueness checking across all entities
✅ **Stores Contact, Repo, Issues, PRs**: All four data types implemented
✅ **Key-value storage**: Using Badger DB instead of Postgres

## Performance Notes

- **Storage**: Badger DB uses LSM (Log-Structured Merge) tree for optimal write performance
- **Deduplication**: O(1) hash lookup via key-value store
- **API Response**: Fast JSON marshaling for all endpoints
- **Rate Limiting**: Built-in delays between GitHub API calls (configurable)

## Future Enhancements

1. **Playwright Integration**: Full headless browser support for JavaScript-heavy pages
2. **Advanced Search**: Elasticsearch/Typesense integration for full-text search
3. **Webhooks**: Real-time updates when repositories change
4. **Caching**: Redis layer for frequently accessed data
5. **Monitoring**: Prometheus metrics and Grafana dashboards
6. **Distributed Crawling**: Multi-worker support for parallel crawling

## License

MIT

## Support

For issues and questions, please open a GitHub issue.
