# Примеры использования API Fyne-on

## 1. Проверка здоровья приложения

```bash
curl -X GET http://localhost:3000/health
```

Ответ:
```json
{
  "status": "ok",
  "message": "Fyne-on crawler is running"
}
```

## 2. Получить статистику базы данных

```bash
curl -X GET http://localhost:3000/stats
```

Ответ:
```json
{
  "repositories": 1250,
  "contacts": 3450,
  "issues": 45230,
  "pull_requests": 12340
}
```

## 3. Запустить crawler

Запустить краулер с начальной точки (username: torvalds):

```bash
curl -X POST http://localhost:3000/crawler/start \
  -H "Content-Type: application/json" \
  -d '{
    "start_username": "torvalds",
    "max_iterations": 10000,
    "delay_ms": 1000,
    "github_token": "your_token_here"
  }'
```

Ответ:
```json
{
  "message": "Crawler started",
  "start_username": "torvalds",
  "max_iterations": 10000,
  "delay_ms": 1000
}
```

С GitHub токеном (повышает лимит запросов):

```bash
curl -X POST http://localhost:3000/crawler/start \
  -H "Content-Type: application/json" \
  -d '{
    "start_username": "gvanrossum",
    "max_iterations": 5000,
    "delay_ms": 500,
    "github_token": "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  }'
```

## 4. Получить все репозитории

```bash
curl -X GET http://localhost:3000/repos
```

## 5. Получить конкретный репозиторий

```bash
curl -X GET http://localhost:3000/repos/torvalds/linux
```

Ответ:
```json
{
  "id": "torvalds/linux",
  "name": "linux",
  "owner": "torvalds",
  "url": "https://github.com/torvalds/linux",
  "description": "Linux kernel source tree",
  "stars": 180000,
  "language": "C",
  "has_open_license": true,
  "license": "GPL-2.0",
  "hash": "sha256hash...",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

## 6. Получить issues репозитория

```bash
curl -X GET http://localhost:3000/repos/torvalds/linux/issues
```

Ответ:
```json
[
  {
    "id": "123456",
    "repo_id": "torvalds/linux",
    "title": "CPU frequency scaling issue",
    "url": "https://github.com/torvalds/linux/issues/123456",
    "state": "open",
    "body": "Issue description...",
    "author": "username",
    "created_at": "2024-01-10T15:20:00Z",
    "updated_at": "2024-01-15T10:30:00Z",
    "hash": "sha256hash..."
  }
]
```

## 7. Получить PRs репозитория

```bash
curl -X GET http://localhost:3000/repos/golang/go/prs
```

## 8. Поиск репозиториев по языку

```bash
curl -X GET "http://localhost:3000/repos/search?language=Go&min_stars=100"
```

## 9. Получить все контакты (разработчики)

```bash
curl -X GET http://localhost:3000/contacts
```

Ответ:
```json
[
  {
    "id": "1",
    "login": "torvalds",
    "url": "https://github.com/torvalds",
    "avatar": "https://avatars.githubusercontent.com/u/1?v=4",
    "company": "Linux Foundation",
    "email": "",
    "location": "Finland",
    "bio": "I'm the original creator of Linux",
    "hash": "sha256hash...",
    "updated_at": "2024-01-15T10:30:00Z"
  }
]
```

## 10. Получить контакт по username

```bash
curl -X GET http://localhost:3000/contacts/torvalds
```

## 11. Удалить репозиторий

```bash
curl -X DELETE http://localhost:3000/repos/owner/repo
```

Ответ:
```json
{
  "message": "repository deleted"
}
```

## 12. Получить список всех маршрутов

```bash
curl -X GET http://localhost:3000/api/routes
```

## Примеры использования с Python

```python
import requests
import json

BASE_URL = "http://localhost:3000"

# Запустить crawler
def start_crawler():
    payload = {
        "start_username": "torvalds",
        "max_iterations": 5000,
        "delay_ms": 1000,
        "github_token": "your_token"
    }
    response = requests.post(f"{BASE_URL}/crawler/start", json=payload)
    print(json.dumps(response.json(), indent=2))

# Получить статистику
def get_stats():
    response = requests.get(f"{BASE_URL}/stats")
    print(json.dumps(response.json(), indent=2))

# Получить репозитории
def get_repos():
    response = requests.get(f"{BASE_URL}/repos")
    repos = response.json()
    print(f"Found {len(repos)} repositories")
    for repo in repos[:5]:  # Show first 5
        print(f"  - {repo['owner']}/{repo['name']} ({repo['stars']} stars)")

# Получить контакты
def get_contacts():
    response = requests.get(f"{BASE_URL}/contacts")
    contacts = response.json()
    print(f"Found {len(contacts)} contacts")
    for contact in contacts[:5]:
        print(f"  - {contact['login']} ({contact['location']})")

if __name__ == "__main__":
    print("=== Starting Crawler ===")
    start_crawler()
    
    print("\n=== Statistics ===")
    get_stats()
    
    print("\n=== Repositories ===")
    get_repos()
    
    print("\n=== Contacts ===")
    get_contacts()
```

## Примеры использования с bash/curl

```bash
#!/bin/bash

BASE_URL="http://localhost:3000"

# Проверить здоровье
echo "=== Health Check ==="
curl -s ${BASE_URL}/health | jq .

# Получить статистику
echo "=== Statistics ==="
curl -s ${BASE_URL}/stats | jq .

# Запустить crawler
echo "=== Starting Crawler ==="
curl -s -X POST ${BASE_URL}/crawler/start \
  -H "Content-Type: application/json" \
  -d '{
    "start_username": "torvalds",
    "max_iterations": 5000,
    "delay_ms": 1000
  }' | jq .

# Проверить прогресс (повторяйте несколько раз)
echo "=== Checking Progress ==="
sleep 10
curl -s ${BASE_URL}/stats | jq .

# Получить top репозитории
echo "=== Top Repositories ==="
curl -s "${BASE_URL}/repos/search?language=C&min_stars=1000" | jq '.[0:5]'

# Получить контакты
echo "=== Contacts ==="
curl -s ${BASE_URL}/contacts | jq '.[0:5]'
```

## Важные замечания

1. **Rate Limiting**: GitHub API имеет лимиты на количество запросов. Используйте `delay_ms` и GitHub токен для оптимизации.

2. **GitHub Token**: Получить токен можно на странице https://github.com/settings/tokens
   - Требуется scopes: `public_repo`, `read:user`

3. **Crawler Progress**: Crawler работает асинхронно. Проверяйте статистику чтобы увидеть прогресс.

4. **Deduplication**: Система автоматически не добавляет дубликаты базываясь на хешах.

5. **Storage**: Данные хранятся в `./badger_data/` локально на диске.

## Мониторинг процесса крауллинга

```bash
# Монитор в реальном времени (обновляется каждые 5 секунд)
watch -n 5 'curl -s http://localhost:3000/stats | jq .'

# Или с jq для красивого вывода
while true; do
  echo "=== $(date) ==="
  curl -s http://localhost:3000/stats | jq .
  sleep 5
done
```
