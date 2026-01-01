# Fyne-on - GitHub Crawler with Markov Chains

High-performance GitHub crawler using Markov chains for intelligent traversal and Badger KV for efficient data storage.

## 🚀 Features

- **Markov Chain-based Crawling**: Intelligent URL traversal using Markov chains
- **Key-Value Storage**: Badger DB for fast and reliable data storage
- **REST API**: Complete REST API for querying and managing collected data
- **Deduplication**: SHA256 hash-based deduplication
- **Scalability**: Architecture scales to 10,000+ repositories and contacts

## 📂 Architecture

### Core Components

1. **BadgerDB** (`pkg/database/badgerdb.go`)
   - Key-value database wrapper around Badger
   - SHA256 hashing for deduplication

2. **GitHub Crawler** (`internal/core/crawler/github.go`)
   - GitHub API and HTML parsing
   - Collects repository, issue, and PR information

3. **Storage Service** (`pkg/storage/storage.go`)
   - CRUD operations for all data types
   - Hash-based deduplication

4. **Markov Chain** (`internal/core/markov/markov.go`)
   - Intelligent URL selection for traversal

## ⚡ Quick Start

```bash
go build -o bin/app.exe ./cmd/app
./bin/app.exe
```

API available at `http://localhost:3000`

## 📊 Data Types

Collects:
- **Repositories**: GitHub repositories
- **Issues**: Issues and tasks
- **Pull Requests**: Code review requests
- **Contacts**: User information

## 🔗 Main Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /health | Health check |
| GET | /stats | Database statistics |
| POST | /crawler/start | Start crawler |
| POST | /export/all-jsonl | Export to JSONL |

## 📝 Data Export

Supported formats:
- JSONL (JSON Lines) - recommended for LLM training
- Parquet - for big data analytics

## 🐳 Docker

```bash
docker build -t fyne-on .
docker run -p 3000:3000 fyne-on
```

## 📄 License

MIT
