# Fyne-on - GitHub Crawler with Markov Chains

High-performance GitHub crawler using Markov chains for intelligent traversal and Badger KV for efficient data storage.

**Documentation**:
- 🇷🇺 [Russian Documentation](./docs/Ru/)
- 🇬🇧 [English Documentation](./docs/Eng/)

## Quick Start

```bash
go build -o bin/app.exe ./cmd/app
./bin/app.exe
```

API available at `http://localhost:3000`

## Features

- Markov Chain-based crawling
- Badger KV storage
- REST API with 12+ endpoints
- JSONL export for LLM training
- Hash-based deduplication
- Scalable to 10,000+ repositories

## Architecture

```
Fyne-on/
├── cmd/app/main.go           # REST API
├── pkg/
│   ├── database/             # Badger KV wrapper
│   ├── models/               # Data structures
│   ├── parquet/              # JSONL export
│   ├── storage/              # Storage service
│   └── scraper/              # Web scraping
├── internal/core/
│   ├── crawler/              # GitHub crawler
│   └── markov/               # Markov chains
└── docs/
    ├── Ru/                   # Russian docs
    └── Eng/                  # English docs
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /health | Health check |
| GET | /stats | Database statistics |
| POST | /crawler/start | Start crawler |
| POST | /export/all-jsonl | Export to JSONL |

## License

MIT
