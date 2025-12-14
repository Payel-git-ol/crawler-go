Понял тебя, Павел 👌 — ты хочешь видеть весь `README.md` как единый файл, в одном большом кодовом блоке, без разрывов. Вот он целиком:

```markdown
# Fyne-on

Backend‑сервис для краулинга GitHub и предоставления данных через REST API.

---

## 🚀 Запуск

```bash
# Установить зависимости
go mod download

# Запустить сервер
go run ./cmd/app

# Проверить работу
curl http://localhost:3000/health
```

Сервер стартует на порту `3000`.

---

## 📂 Архитектура проекта

```
Fyne-on/
├── cmd/app/main.go        # REST API + роутер (Fiber)
├── pkg/
│   ├── crawler/github.go  # Краулер GitHub API + HTML
│   ├── database/          # Обертка над Badger KV
│   ├── models/models.go   # Структуры данных (Contact, Repo, Issue, PR)
│   ├── scraper/http.go    # Утилиты для web scraping
│   └── storage/storage.go # Storage Service (CRUD + deduplication)
├── Dockerfile
├── docker-compose.yaml
├── go.mod
├── README.md
└── ...
```

---

## 🔗 API эндпоинты

### Health & Stats
- `GET /health` — проверка состояния
- `GET /stats` — статистика по БД
- `GET /stats/summary` — компактные счётчики

### Repositories
- `GET /repos` — список репозиториев  
  Параметры:
    - `expand=true` — расширенные поля
    - `include_issues=count` — добавить количество issues
- `GET /repos/:owner/:name` — конкретный репозиторий
- `GET /repos/:owner/:name/issues` — issues репозитория
- `GET /repos/:owner/:name/prs` — pull requests репозитория
- `GET /repos/search?language=Go` — поиск по языку
- `DELETE /repos/:owner/:name` — удалить репозиторий

### Issues
- `GET /issues?page=1&limit=100` — постраничный список всех issues

### Contacts
- `GET /contacts` — список контактов
- `GET /contacts/:login` — конкретный контакт

### Crawler
- `POST /crawler/start` — запустить краулер  
  Тело запроса:
  ```json
  {
    "start_usernames": ["microsoft", "google"],
    "max_iterations": 20000,
    "delay_ms": 1000,
    "github_token": "YOUR_TOKEN",
    "use_playwright": true
  }
  ```
    Или для HTML скрапинга 
```json
    {
  "start_usernames": [
    "microsoft",
    "google",
    "facebook",
    "apache",
    "mozilla",
    "aws",
    "tensorflow",
    "kubernetes",
    "apple",
    "oracle",
    "rust-lang",
    "golang",
    "python",
    "django",
    "spring-projects",
    "dotnet",
    "linux",
    "debian",
    "homebrew",
    "kubernetes-sigs",
    "apache-spark",
    "gnome",
    "qt",
    "openai",
    "facebookresearch",
    "googleapis",
    "huggingface",
    "pytorch",
    "hashicorp",
    "helm",
    "ansible",
    "jenkinsci",
    "grafana",
    "prometheus",
    "mongodb",
    "cockroachdb",
    "neo4j",
    "redis",
    "elastic",
    "apache", "apache-spark", "apache-flink", "apache-kafka",
    "cncf", "kubernetes-sigs", "helm", "istio", "linkerd",
    "hashicorp", "terraform-providers", "ansible", "chef",
    "grafana", "prometheus", "influxdata",
    "elastic", "opensearch-project",
    "redis", "memcached",
    "postgres", "mysql", "sqlite",
    "rust-lang", "golang", "python", "django", "numpy", "scipy", "pandas-dev",
    "huggingface", "pytorch", "tensorflow", "openai",
    "mozilla", "gnome", "qt", "electron", "vercel", "netlify","numpy", "scipy", "pandas-dev", "matplotlib", "scikit-learn",
    "electron", "vercel", "netlify", "nextjs", "gatsbyjs",
    "ansible", "chef", "puppetlabs", "saltstack",
    "influxdata", "timescale", "vitessio",
    "opensearch-project", "apache-flink", "apache-kafka",
    "cncf", "istio", "linkerd"

  ],
  "delay_ms": 1000,
  "use_playwright": true
}

```  

- `GET /crawler/config` — текущая конфигурация краулера

### Service
- `GET /api/routes` — список всех маршрутов

---

## 📖 Примеры использования

```bash
# Проверка состояния
curl http://localhost:3000/health

# Получить статистику
curl http://localhost:3000/stats/summary

# Список репозиториев (с расширенными полями)
curl "http://localhost:3000/repos?expand=true"

# Репозиторий Python/cpython
curl http://localhost:3000/repos/python/cpython

# Issues репозитория
curl http://localhost:3000/repos/python/cpython/issues

# Все issues постранично
curl "http://localhost:3000/issues?page=1&limit=50"

# Запуск краулера
curl -X POST http://localhost:3000/crawler/start \
  -H "Content-Type: application/json" \
  -d '{"start_usernames":["microsoft"],"max_iterations":5000,"delay_ms":500,"use_playwright":true}'
```

---

## 🧪 Тестирование

```bash
go test ./... -v
```

---

## ⚡ Troubleshooting

- **GitHub rate limit** → используйте токен (`github_token`)
- **Большие ответы `/issues`** → используйте `page` и `limit`
- **Badger LOCK file** → удалите `badger_data/LOCK`

---

## 📦 Docker

```bash
docker build -t fyne-on:latest .
docker run -p 3000:3000 fyne-on:latest
```

---

## 🛠 Code Style

- Используйте `gofmt` для форматирования
- Используйте `golint` для проверки стиля
- Названия переменных в `camelCase`
- Константы в `UPPER_CASE`
- Экспортируемые функции комментируйте

```bash
gofmt -w .
golangci-lint run ./...
```

---

## 🔬 Профилирование

### CPU profiling

```go
import _ "net/http/pprof"

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
```

---

Теперь у тебя полный `README.md` в одном большом кодовом блоке, как файл.