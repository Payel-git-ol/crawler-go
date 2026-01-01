# Getting Started with Data Collection

## 3 Ways to Run Crawler

### Method 1: Batch Run (RECOMMENDED)

```powershell
powershell -ExecutionPolicy Bypass -File start_crawler_batch.ps1
```

Runs crawler for multiple companies, shows live statistics, and exports to JSONL.

### Method 2: Single Company

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

### Method 3: Multiple Companies in Loop

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

## Check Statistics

```powershell
$stats = Invoke-WebRequest -Uri "http://localhost:3000/stats" -UseBasicParsing | ConvertFrom-Json
Write-Host "Repositories:  $($stats.repositories)"
Write-Host "Issues:        $($stats.issues)"
Write-Host "Pull Requests: $($stats.pull_requests)"
```

## Real-time Monitoring

```powershell
while ($true) {
    Clear-Host
    $stats = Invoke-WebRequest -Uri "http://localhost:3000/stats" -UseBasicParsing | ConvertFrom-Json
    Write-Host "📦 Repositories:  $($stats.repositories)"
    Write-Host "📝 Issues:        $($stats.issues)"
    Write-Host "🔀 Pull Requests: $($stats.pull_requests)"
    Start-Sleep -Seconds 10
}
```

## Export to JSONL

```powershell
$export = Invoke-WebRequest -Uri "http://localhost:3000/export/all-jsonl" -Method POST -UseBasicParsing | ConvertFrom-Json
Write-Host "Issues:        $($export.issues_count)"
Write-Host "Pull Requests: $($export.pull_requests_count)"
Write-Host "Repositories:  $($export.repositories_count)"
```

## Check Generated Files

```powershell
Get-ChildItem ./jsonl_data/ -Filter "*.jsonl" | Select-Object Name, @{N="Size (MB)";E={[math]::Round($_.Length/1MB,2)}}
```
