$uri = "http://localhost:3000/crawler/start"
$json = '{"StartUsername":"torvalds","MaxIter":20,"DelayMs":1000}'

# Используем Invoke-WebRequest с параметрами
try {
    $result = Invoke-WebRequest `
        -Uri $uri `
        -Method POST `
        -ContentType "application/json" `
        -Body $json `
        -UseBasicParsing `
        -TimeoutSec 300

    Write-Host "Status: $($result.StatusCode)"
    $result.Content
}
catch {
    Write-Host "Error: $($_)"
    Write-Host "Exception: $($_.Exception.Message)"
}
