# Быстрый старт: Экспорт для LLM

## За 3 шага

### Шаг 1: Запустить приложение
```bash
cd Fyne-on
./bin/app.exe
```

### Шаг 2: Запустить краулер GitHub
```powershell
$json = @{
    start_usernames = @("microsoft")
    max_iterations = 5000
    delay_ms = 500
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:3000/crawler/start" `
    -Method POST `
    -ContentType "application/json" `
    -Body $json `
    -UseBasicParsing
```

### Шаг 3: Экспортировать в JSONL
```powershell
Invoke-WebRequest -Uri "http://localhost:3000/export/all-jsonl" -Method POST -UseBasicParsing
```

## Результат

Файлы появятся в `./jsonl_data/`:
- `issues_TIMESTAMP.jsonl` - Issues (рекомендуется для LLM)
- `pull_requests_TIMESTAMP.jsonl` - Pull requests
- `repositories_TIMESTAMP.jsonl` - Репозитории

## Формат JSONL

Каждая строка - валидный JSON объект:
```json
{"id": "123", "title": "Bug fix", "body": "...", "state": "closed"}
{"id": "124", "title": "Feature", "body": "...", "state": "open"}
```

Идеально для машинного обучения!

## Использование в Python

```python
import pandas as pd

df = pd.read_json('./jsonl_data/issues_*.jsonl', lines=True)
print(f"Загружено {len(df)} issues")
```

## Использование в PyTorch

```python
from torch.utils.data import DataLoader, Dataset
import json

class IssuesDataset(Dataset):
    def __init__(self, jsonl_file):
        with open(jsonl_file) as f:
            self.data = [json.loads(line) for line in f]
    
    def __getitem__(self, idx):
        return self.data[idx]
    
    def __len__(self):
        return len(self.data)

dataset = IssuesDataset('./jsonl_data/issues_*.jsonl')
loader = DataLoader(dataset, batch_size=32)
```
