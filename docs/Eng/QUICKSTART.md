# Quick Start: Export for LLM

## In 3 Steps

### Step 1: Run Application
```bash
cd Fyne-on
./bin/app.exe
```

### Step 2: Start GitHub Crawler
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

### Step 3: Export to JSONL
```powershell
Invoke-WebRequest -Uri "http://localhost:3000/export/all-jsonl" -Method POST -UseBasicParsing
```

## Results

Files created in `./jsonl_data/`:
- `issues_TIMESTAMP.jsonl` - Issues (recommended for LLM)
- `pull_requests_TIMESTAMP.jsonl` - Pull requests
- `repositories_TIMESTAMP.jsonl` - Repositories

## JSONL Format

Each line is a valid JSON object:
```json
{"id": "123", "title": "Bug fix", "body": "...", "state": "closed"}
{"id": "124", "title": "Feature", "body": "...", "state": "open"}
```

Perfect for machine learning!

## Python Usage

```python
import pandas as pd

df = pd.read_json('./jsonl_data/issues_*.jsonl', lines=True)
print(f"Loaded {len(df)} issues")
```

## PyTorch Usage

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
