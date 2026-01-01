# ============================================================
# Fyne-on: Массовый запуск краулера для нескольких компаний
# ============================================================

# Настройки
$ApiUrl = "http://localhost:3000"
$Companies = @(
    "microsoft",
    "google",
    "amazon",
    "apple",
    "facebook",
    "uber",
    "airbnb"
)

# Параметры краулера
$MaxIterations = 5000      # Сколько репозиториев собрать
$DelayMs = 500             # Задержка между запросами (мс)

Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "🚀 FYNE-ON КРАУЛЕР - МАССОВЫЙ ЗАПУСК" -ForegroundColor Yellow
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

# 1. Проверка что приложение работает
Write-Host "📡 Проверка подключения к API..." -ForegroundColor Blue
try {
    $health = Invoke-WebRequest -Uri "$ApiUrl/health" -UseBasicParsing
    if ($health.StatusCode -eq 200) {
        Write-Host "✅ API доступен!" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ API недоступен на $ApiUrl" -ForegroundColor Red
    Write-Host "💡 Запусти приложение сначала: .\bin\app.exe" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "📋 ПАРАМЕТРЫ СБОРА:" -ForegroundColor Blue
Write-Host "  • Компании:      $($Companies.Count) штук"
Write-Host "  • Макс. итераций: $MaxIterations"
Write-Host "  • Задержка:      $DelayMs мс"
Write-Host ""

# 2. Запуск краулера для каждой компании
$StartTime = Get-Date
$Counter = 1

foreach ($company in $Companies) {
    Write-Host "[$Counter/$($Companies.Count)] Запуск для: $company" -ForegroundColor Magenta
    
    $body = @{
        start_usernames = @($company)
        max_iterations = $MaxIterations
        delay_ms = $DelayMs
    } | ConvertTo-Json
    
    try {
        $response = Invoke-WebRequest -Uri "$ApiUrl/crawler/start" `
            -Method POST `
            -ContentType "application/json" `
            -Body $body `
            -UseBasicParsing
        
        $result = $response.Content | ConvertFrom-Json
        Write-Host "   ├─ Status: $($result.message)" -ForegroundColor Green
        Write-Host "   └─ ID: $($result.crawler_id)" -ForegroundColor Green
    } catch {
        Write-Host "   ❌ Ошибка при запуске для $company" -ForegroundColor Red
        Write-Host "   └─ Error: $_" -ForegroundColor Red
    }
    
    $Counter++
    Write-Host ""
}

# 3. Мониторинг статистики
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "📊 МОНИТОРИНГ СБОРА ДАННЫХ" -ForegroundColor Yellow
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

$MonitoringTime = 0
$MaxMonitoringTime = 600  # 10 минут мониторинга

while ($MonitoringTime -lt $MaxMonitoringTime) {
    try {
        $stats = Invoke-WebRequest -Uri "$ApiUrl/stats" `
            -UseBasicParsing | ConvertFrom-Json
        
        $elapsed = ((Get-Date) - $StartTime).TotalMinutes
        
        Write-Host "[$(Get-Date -Format 'HH:mm:ss')] Прогресс сбора:" -ForegroundColor Cyan
        Write-Host "  📦 Репозитории:   $($stats.repositories_count)" -ForegroundColor Green
        Write-Host "  📝 Issues:        $($stats.issues_count)" -ForegroundColor Green
        Write-Host "  🔀 Pull Requests: $($stats.pull_requests_count)" -ForegroundColor Green
        Write-Host "  👥 Контакты:      $($stats.contacts_count)" -ForegroundColor Green
        Write-Host "  ⏱️  Прошло:        $([math]::Round($elapsed, 1)) минут" -ForegroundColor Yellow
        Write-Host ""
        
        # Пауза перед следующей проверкой
        Start-Sleep -Seconds 10
        $MonitoringTime += 10
    } catch {
        Write-Host "⚠️  Не удалось получить статистику" -ForegroundColor Yellow
    }
}

# 4. Финальная статистика
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "✅ СБОР ЗАВЕРШЁН" -ForegroundColor Green
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan

$finalStats = Invoke-WebRequest -Uri "$ApiUrl/stats" -UseBasicParsing | ConvertFrom-Json
$totalTime = ((Get-Date) - $StartTime).TotalMinutes

Write-Host ""
Write-Host "📊 ФИНАЛЬНАЯ СТАТИСТИКА:" -ForegroundColor Green
Write-Host "  📦 Репозитории:   $($finalStats.repositories_count)" -ForegroundColor White
Write-Host "  📝 Issues:        $($finalStats.issues_count)" -ForegroundColor White
Write-Host "  🔀 Pull Requests: $($finalStats.pull_requests_count)" -ForegroundColor White
Write-Host "  👥 Контакты:      $($finalStats.contacts_count)" -ForegroundColor White
Write-Host "  ⏱️  Общее время:   $([math]::Round($totalTime, 1)) минут" -ForegroundColor White
Write-Host ""

# 5. Экспорт в JSONL
Write-Host "📤 ЭКСПОРТ В JSONL..." -ForegroundColor Blue
$exportStart = Get-Date

try {
    $exportResponse = Invoke-WebRequest -Uri "$ApiUrl/export/all-jsonl" `
        -Method POST `
        -UseBasicParsing
    
    $exportResult = $exportResponse.Content | ConvertFrom-Json
    $exportTime = ((Get-Date) - $exportStart).TotalSeconds
    
    Write-Host "✅ Экспорт успешен!" -ForegroundColor Green
    Write-Host "  📁 Issues:        $($exportResult.issues_count) строк" -ForegroundColor Green
    Write-Host "  📁 Pull Requests: $($exportResult.pull_requests_count) строк" -ForegroundColor Green
    Write-Host "  📁 Repositories:  $($exportResult.repositories_count) строк" -ForegroundColor Green
    Write-Host "  ⏱️  Время экспорта: $([math]::Round($exportTime, 1)) сек" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "⚠️  Ошибка при экспорте: $_" -ForegroundColor Yellow
}

# 6. Показать где находятся файлы
Write-Host "📂 ФАЙЛЫ НАХОДЯТСЯ ТУТ:" -ForegroundColor Blue
$jsonlFiles = Get-ChildItem -Path "./jsonl_data" -Filter "*.jsonl" -ErrorAction SilentlyContinue
if ($jsonlFiles) {
    foreach ($file in $jsonlFiles) {
        $sizeKB = [math]::Round($file.Length / 1KB, 2)
        $lines = @(Get-Content $file.FullName).Count
        Write-Host "  📄 $($file.Name) ($sizeKB KB, $lines строк)" -ForegroundColor Green
    }
} else {
    Write-Host "  ⚠️  JSONL файлы не найдены" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "🎉 ВСЕ ГОТОВО! ДАННЫЕ ЭКСПОРТИРОВАНЫ В JSONL" -ForegroundColor Green
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""
Write-Host "💡 Что дальше?" -ForegroundColor Yellow
Write-Host "   1. Используй JSONL файлы для обучения LLM" -ForegroundColor White
Write-Host "   2. Смотри примеры в USAGE_EXAMPLES.md" -ForegroundColor White
Write-Host "   3. Прочитай LLM_GUIDE_WINDOWS.md для интеграции" -ForegroundColor White
Write-Host ""
