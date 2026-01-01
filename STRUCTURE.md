PROJECT STRUCTURE
=================

Fyne-on/
├── README_MAIN.md              # Main README with doc links
│
├── bin/
│   └── app.exe                 # Compiled binary
│
├── cmd/
│   ├── app/main.go             # REST API server
│   ├── debug/main.go           # Debug utilities
│   └── seed/main.go            # Database seeding
│
├── pkg/
│   ├── database/
│   │   ├── badgerdb.go         # Badger KV wrapper
│   │   └── badgerdb_test.go    # Database tests
│   ├── models/models.go        # Data structures
│   ├── parquet/exporter.go     # JSONL export
│   ├── storage/storage.go      # CRUD service
│   └── scraper/http.go         # Web scraping
│
├── internal/core/
│   ├── crawler/
│   │   ├── github.go           # GitHub crawler
│   │   └── telegram.go         # Telegram crawler
│   └── markov/
│       ├── markov.go           # Markov chains
│       └── markov_test.go      # Chain tests
│
├── docs/
│   ├── Ru/                     # Russian documentation
│   │   ├── README.md
│   │   ├── QUICKSTART.md
│   │   └── START_CRAWLING.md
│   │
│   └── Eng/                    # English documentation
│       ├── README.md
│       ├── QUICKSTART.md
│       └── START_CRAWLING.md
│
├── jsonl_data/                 # Exported JSONL files
│   ├── issues_*.jsonl
│   ├── pull_requests_*.jsonl
│   └── repositories_*.jsonl
│
├── badger_data/                # Database storage
│
├── config.yaml                 # Configuration file
├── docker-compose.yaml         # Docker compose
├── Dockerfile                  # Container image
├── go.mod                      # Go dependencies
├── LICENSE                     # MIT License
└── Makefile                    # Build commands

DOCUMENTATION STRUCTURE
=======================

docs/Ru/ (Русская документация)
  ├── README.md          - Обзор проекта
  ├── QUICKSTART.md      - Быстрый старт
  └── START_CRAWLING.md  - Запуск сбора данных

docs/Eng/ (English documentation)
  ├── README.md          - Project overview
  ├── QUICKSTART.md      - Quick start guide
  └── START_CRAWLING.md  - Data collection guide

MICROSERVICE ARCHITECTURE
========================

The project is organized as a microservice:
- Single responsibility principle
- Clean separation of concerns
- Modular packages
- Easy to extend and maintain

Key packages:
- database: KV storage abstraction
- models: Data structures
- storage: CRUD operations
- crawler: Data collection
- parquet: Data export
