# Fyne-on Development Guide

## Архитектура проекта

### Структура папок

```
Fyne-on/
├── cmd/
│   └── app/
│       └── main.go              # Точка входа приложения
├── pkg/
│   ├── crawler/                 # GitHub краулер
│   │   ├── github.go            # Основная логика крауллинга
│   │   └── http.go              # HTTP утилиты (для Playwright)
│   ├── database/                # Badger KV обертка
│   │   └── badgerdb.go          # Реализация DB операций
│   ├── markov/                  # Markov chain логика
│   │   └── markov.go            # Вероятностные переходы
│   ├── models/                  # Данные модели
│   │   └── models.go            # Contact, Repo, Issue, PR
│   ├── scraper/                 # Web scraper
│   │   └── http.go              # HTTP скрепер утилиты
│   └── storage/                 # Слой хранилища
│       └── storage.go           # CRUD операции
├── docker-compose.yaml          # Docker конфигурация
├── go.mod / go.sum              # Go модули
└── README.md                    # Документация
```

## Добавление новой функции

### 1. Новый тип данных

Добавьте структуру в `pkg/models/models.go`:

```go
type NewEntity struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Hash      string    `json:"hash"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### 2. Методы хранилища

Добавьте в `pkg/storage/storage.go`:

```go
func (s *StorageService) SaveNewEntity(entity models.NewEntity) (bool, error) {
    key := "new_entity:" + entity.ID
    
    if entity.Hash == "" {
        entity.Hash = database.GenerateHash(entity.ID, entity.Name)
    }
    
    entity.UpdatedAt = time.Now()
    return true, s.db.Set(key, entity)
}

func (s *StorageService) GetNewEntity(id string) (*models.NewEntity, error) {
    key := "new_entity:" + id
    var entity models.NewEntity
    err := s.db.GetJSON(key, &entity)
    if err != nil {
        return nil, err
    }
    return &entity, nil
}
```

### 3. Функция крауллинга

Добавьте в `pkg/crawler/github.go`:

```go
func (gc *GithubCrawler) FetchNewEntities(param string) ([]models.NewEntity, error) {
    url := fmt.Sprintf("https://api.github.com/...")
    body, err := gc.makeRequest(url)
    if err != nil {
        return nil, err
    }
    
    var data []struct {
        // JSON fields
    }
    
    if err := json.Unmarshal(body, &data); err != nil {
        return nil, err
    }
    
    entities := []models.NewEntity{}
    for _, d := range data {
        entity := models.NewEntity{
            // Populate fields
        }
        entities = append(entities, entity)
    }
    
    return entities, nil
}
```

### 4. REST API endpoint

Добавьте в `cmd/app/main.go`:

```go
// Get all new entities
app.Get("/new-entities", func(c fiber.Ctx) error {
    // Implementation
    return c.JSON(fiber.Map{})
})

// Get specific new entity
app.Get("/new-entities/:id", func(c fiber.Ctx) error {
    id := c.Params("id")
    entity, err := storageService.GetNewEntity(id)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "not found"})
    }
    return c.JSON(entity)
})
```

## Запуск в Development режиме

```bash
# 1. Установить зависимости
go mod download

# 2. Запустить в debug режиме
go run ./cmd/app

# 3. В другом терминале проверить API
curl http://localhost:3000/health
```

## Testing

### Unit тесты

```bash
go test ./... -v
```

### Coverage

```bash
go test ./... -cover
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration тесты

```go
// Пример в pkg/crawler/github_test.go
func TestFetchUserProfile(t *testing.T) {
    crawler := NewGithubCrawler(nil)
    contact, err := crawler.FetchUserProfile("torvalds")
    assert.NoError(t, err)
    assert.NotNil(t, contact)
    assert.Equal(t, "torvalds", contact.Login)
}
```

## Отладка

### Логирование

```go
import "log"

log.Printf("Debug: %v\n", value)
log.Fatalf("Error: %v\n", err)
```

### Инспекция БД

```bash
# Использовать badger CLI tools для инспекции
# Или написать простой скрипт для чтения данных
```

## Performance Optimization

### 1. Кэширование

```go
// Добавить Redis кэш слой
type CachedStorage struct {
    db    *database.BadgerDB
    cache redis.Client
}
```

### 2. Batch операции

```go
// Вместо одиночного сохранения
func (s *StorageService) SaveBatch(entities []models.Repo) error {
    return s.db.db.Batch(func(txn *badger.Txn) error {
        for _, entity := range entities {
            // Save each entity
        }
        return nil
    })
}
```

### 3. Индексирование

```go
// Добавить вторичные индексы
key := fmt.Sprintf("repo:lang:%s:%s/%s", language, owner, name)
s.db.Set(key, repoID)
```

## Миграция данных

### Экспорт

```bash
# Экспортировать из Badger
curl -X POST http://localhost:3000/export > data.json
```

### Импорт

```bash
# Импортировать в Badger
curl -X POST http://localhost:3000/import \
  -H "Content-Type: application/json" \
  -d @data.json
```

## Обработка ошибок

### Best practices

```go
// Не делайте так
if err != nil {
    panic(err)
}

// Делайте так
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Логируйте контекст
log.Printf("Failed to fetch %s: %v", url, err)
```

## Code Style

Следуйте Go conventions:
- Используйте `gofmt` для форматирования
- Используйте `golint` для проверки стиля
- Названия переменных в camelCase
- Названия констант в UPPER_CASE
- Комментируйте exported функции

```bash
gofmt -w .
golangci-lint run ./...
```

## Git workflow

```bash
# Создать feature branch
git checkout -b feature/new-feature

# Коммиты
git commit -m "feat: add new feature"

# Push
git push origin feature/new-feature

# Pull request
# Описать изменения
# Получить approval
# Merge в main
```

## Deployment

### Build для продакшена

```bash
# Cross-compile для Linux
GOOS=linux GOARCH=amd64 go build -o bin/app ./cmd/app

# Оптимизированная сборка
go build -ldflags="-s -w" -o bin/app ./cmd/app
```

### Docker image

```bash
docker build -t fyne-on:latest .
docker run -p 3000:3000 fyne-on:latest
```

## Troubleshooting

### Issue: Badger LOCK file

```bash
# Решение: удалить LOCK файл
rm -f badger_data/LOCK
```

### Issue: GitHub rate limit

```bash
# Решение: использовать token с более высоким лимитом
# Или уменьшить delay_ms
```

### Issue: High memory usage

```bash
# Решение: уменьшить max_iterations
# Или запускать в несколько сессий
```

## Профилирование

### CPU profiling

```go
import _ "net/http/pprof"

// Запуск pprof сервера
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

```bash
go tool pprof http://localhost:6060/debug/pprof/profile
```

### Memory profiling

```bash
go tool pprof http://localhost:6060/debug/pprof/heap
```

## Документация

### Генерирование документации

```bash
# Установить godoc
go install golang.org/x/tools/cmd/godoc@latest

# Запустить локально
godoc -http=:6060

# Посетить http://localhost:6060
```

### Комментирование

```go
// Package crawler implements GitHub repository crawling
package crawler

// GithubCrawler represents a GitHub crawler instance
type GithubCrawler struct {
    // ... fields
}

// FetchUserProfile fetches a GitHub user's profile
// It makes an HTTP request to GitHub API and returns the user data
func (gc *GithubCrawler) FetchUserProfile(username string) (*models.Contact, error) {
    // ... implementation
}
```
