# ==========================================
# Fyne-on: Экспорт данных для LLM обучения
# ==========================================

# 1. Проверить что приложение запущено
Write-Host "✓ Проверка здоровья приложения..." -ForegroundColor Green
$health = Invoke-WebRequest -Uri "http://localhost:3000/health" -UseBasicParsing
$health | ConvertFrom-Json | Format-Table -AutoSize

# 2. Запустить краулер GitHub
Write-Host "`n✓ Запуск GitHub краулера..." -ForegroundColor Green
Write-Host "  (это может занять время в зависимости от max_iterations)"

$crawlJson = @{
    start_username = "torvalds"
    max_iterations = 100
    delay_ms = 1000
} | ConvertTo-Json

$crawlResponse = Invoke-WebRequest `
    -Uri "http://localhost:3000/crawler/start" `
    -Method POST `
    -ContentType "application/json" `
    -Body $crawlJson `
    -UseBasicParsing

$crawlResponse.Content | ConvertFrom-Json | Format-Table -AutoSize

# 3. Ждём немного пока краулер работает
Write-Host "`n⏳ Ожидание завершения краулера... (30 сек)" -ForegroundColor Yellow
Start-Sleep -Seconds 30

# 4. Получить статистику
Write-Host "`n✓ Статистика базы данных:" -ForegroundColor Green
$statsResponse = Invoke-WebRequest -Uri "http://localhost:3000/stats/summary" -UseBasicParsing
$statsResponse.Content | ConvertFrom-Json | Format-Table -AutoSize

# 5. Экспортировать в JSONL для LLM обучения
Write-Host "`n✓ Экспорт данных в JSONL формат (для LLM)..." -ForegroundColor Green

$exportResponse = Invoke-WebRequest `
    -Uri "http://localhost:3000/export/all-jsonl" `
    -Method POST `
    -UseBasicParsing

$exportResult = $exportResponse.Content | ConvertFrom-Json
$exportResult | Format-Table -AutoSize

Write-Host "`n✓ Экспорт завершен!" -ForegroundColor Green
Write-Host "  Файлы находятся в: ./jsonl_data/" -ForegroundColor Cyan

# 6. Показать список файлов
Write-Host "`n✓ Файлы для LLM обучения:" -ForegroundColor Green
Get-ChildItem -Path ./jsonl_data -Recurse | Format-Table -Property Name, Length

Write-Host "`n" -ForegroundColor Green
Write-Host "🎉 Готово! Данные экспортированы и готовы для обучения LLM" -ForegroundColor Green
Write-Host "`n📖 Подробнее в PARQUET_GUIDE.md и QUICKSTART_PARQUET.md" -ForegroundColor Cyan
