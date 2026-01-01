# Как начать сбор данных

## 3 способа запустить краулер

### Способ 1: Массовый запуск (РЕКОМЕНДУЕТСЯ)

```powershell
powershell -ExecutionPolicy Bypass -File start_crawler_batch.ps1
```

Запустит краулер для нескольких компаний, покажет live статистику и экспортирует в JSONL.

### Способ 2: Одна компания

```powershell
$body = @{
    start_usernames = @("torvalds")
    max_iterations = 5000
    delay_ms = 500
} | ConvertTo-Json

Invoke-WebRequest -Uri "http://localhost:3000/crawler/start" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body `
    -UseBasicParsing
```

### Способ 3: Несколько компаний в цикле

```powershell
$companies = @("microsoft", "google", "amazon", "apple")
$apiUrl = "http://localhost:3000"

foreach ($company in $companies) {
    $body = @{
        start_usernames = @($company)
        max_iterations = 5000
        delay_ms = 500
    } | ConvertTo-Json
    
    Invoke-WebRequest -Uri "$apiUrl/crawler/start" `
        -Method POST `
        -ContentType "application/json" `
        -Body $body `
        -UseBasicParsing
}
```

## Проверка статистики

```powershell
$stats = Invoke-WebRequest -Uri "http://localhost:3000/stats" -UseBasicParsing | ConvertFrom-Json
Write-Host "Репозитории:   $($stats.repositories)"
Write-Host "Issues:        $($stats.issues)"
Write-Host "Pull Requests: $($stats.pull_requests)"
```

## Мониторинг в реальном времени

```powershell
while ($true) {
    Clear-Host
    $stats = Invoke-WebRequest -Uri "http://localhost:3000/stats" -UseBasicParsing | ConvertFrom-Json
    Write-Host "📦 Репозитории:   $($stats.repositories)"
    Write-Host "📝 Issues:        $($stats.issues)"
    Write-Host "🔀 Pull Requests: $($stats.pull_requests)"
    Start-Sleep -Seconds 10
}
```

## Экспорт в JSONL

```powershell
$export = Invoke-WebRequest -Uri "http://localhost:3000/export/all-jsonl" -Method POST -UseBasicParsing | ConvertFrom-Json
Write-Host "Issues:        $($export.issues_count)"
Write-Host "Pull Requests: $($export.pull_requests_count)"
Write-Host "Repositories:  $($export.repositories_count)"
```

## Проверка файлов

```powershell
Get-ChildItem ./jsonl_data/ -Filter "*.jsonl" | Select-Object Name, @{N="Size (MB)";E={[math]::Round($_.Length/1MB,2)}}
```
