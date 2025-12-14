Отлично, Павел 👌. Я собрал для тебя полный актуальный `EXAMPLES.md` целиком, в одном кодовом блоке, чтобы можно было сразу заменить файл:

```markdown
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

---

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

---

## 3. Получить компактную статистику

```bash
curl -X GET http://localhost:3000/stats/summary
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

---

## 4. Запустить crawler

```bash
curl -X POST http://localhost:3000/crawler/start \
  -H "Content-Type: application/json" \
  -d '{
    "start_usernames": ["torvalds"],
    "max_iterations": 10000,
    "delay_ms": 1000,
    "github_token": "your_token_here",
    "use_playwright": false
  }'
```

Ответ:
```json
{
  "message": "Crawler started (API mode)",
  "start_username": ["torvalds"],
  "max_iterations": 10000,
  "delay_ms": 1000,
  "use_playwright": false
}
```

---

## 5. Получить все репозитории

```bash
curl -X GET http://localhost:3000/repos
```

С параметрами:
- `expand=true` — расширенные поля
- `include_issues=count` — добавить количество issues

---

## 6. Получить конкретный репозиторий

```bash
curl -X GET http://localhost:3000/repos/torvalds/linux
```

---

## 7. Получить issues репозитория

```bash
curl -X GET http://localhost:3000/repos/torvalds/linux/issues
```

---

## 8. Получить PRs репозитория

```bash
curl -X GET http://localhost:3000/repos/golang/go/prs
```

---

## 9. Поиск репозиториев по языку

```bash
curl -X GET "http://localhost:3000/repos/search?language=Go"
```

---

## 10. Получить все контакты (разработчики)

```bash
curl -X GET http://localhost:3000/contacts
```

---

## 11. Получить контакт по username

```bash
curl -X GET http://localhost:3000/contacts/torvalds
```

---

## 12. Удалить репозиторий

```bash
curl -X DELETE http://localhost:3000/repos/owner/repo
```

---

## 13. Получить список всех маршрутов

```bash
curl -X GET http://localhost:3000/api/routes
```

---

## 14. Получить все issues постранично

```bash
curl -X GET "http://localhost:3000/issues?page=1&limit=50"
```

---

## Примеры использования с Python

```python
import requests
import json

BASE_URL = "http://localhost:3000"

def start_crawler():
    payload = {
        "start_usernames": ["torvalds"],
        "max_iterations": 5000,
        "delay_ms": 1000,
        "github_token": "your_token"
    }
    response = requests.post(f"{BASE_URL}/crawler/start", json=payload)
    print(json.dumps(response.json(), indent=2))

def get_stats():
    response = requests.get(f"{BASE_URL}/stats")
    print(json.dumps(response.json(), indent=2))

def get_repos():
    response = requests.get(f"{BASE_URL}/repos?expand=true")
    repos = response.json()
    print(f"Found {len(repos)} repositories")
    for repo in repos[:5]:
        print(f"  - {repo['owner']}/{repo['name']}")

def get_contacts():
    response = requests.get(f"{BASE_URL}/contacts")
    contacts = response.json()
    print(f"Found {len(contacts)} contacts")
    for contact in contacts[:5]:
        print(f"  - {contact['login']}")

if __name__ == "__main__":
    start_crawler()
    get_stats()
    get_repos()
    get_contacts()
```

---

## Примеры использования с bash/curl

```bash
#!/bin/bash

BASE_URL="http://localhost:3000"

echo "=== Health Check ==="
curl -s ${BASE_URL}/health | jq .

echo "=== Statistics ==="
curl -s ${BASE_URL}/stats | jq .

echo "=== Starting Crawler ==="
curl -s -X POST ${BASE_URL}/crawler/start \
  -H "Content-Type: application/json" \
  -d '{
    "start_usernames": ["torvalds"],
    "max_iterations": 5000,
    "delay_ms": 1000
  }' | jq .

echo "=== Checking Progress ==="
sleep 5
curl -s ${BASE_URL}/stats | jq .

echo "=== Top Repositories ==="
curl -s "${BASE_URL}/repos/search?language=C" | jq '.[0:5]'

echo "=== Contacts ==="
curl -s ${BASE_URL}/contacts | jq '.[0:5]'
```

---

## Важные замечания

1. **Rate Limiting**: GitHub API имеет лимиты на количество запросов. Используйте `delay_ms` и GitHub токен.
2. **GitHub Token**: Получить токен можно на https://github.com/settings/tokens (scopes: `public_repo`, `read:user`).
3. **Crawler Progress**: краулер работает асинхронно, проверяйте `/stats`.
4. **Deduplication**: дубликаты автоматически отбрасываются по хешам.
5. **Storage**: данные хранятся в `./badger_data/`.

---

## Мониторинг процесса крауллинга

```bash
watch -n 5 'curl -s http://localhost:3000/stats | jq .'

while true; do
  echo "=== $(date) ==="
  curl -s http://localhost:3000/stats | jq .
  sleep 5
done
```
```

Теперь `EXAMPLES.md` полностью соответствует твоему актуальному API.