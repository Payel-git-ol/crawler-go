# Запуск краулера

$json = @{
    StartUsername = "torvalds"
    MaxIter = 30
    DelayMs = 500
} | ConvertTo-Json

# Используем .NET для запроса
Add-Type -AssemblyName System.Net.Http

$client = New-Object System.Net.Http.HttpClient
$content = New-Object System.Net.Http.StringContent($json, [System.Text.Encoding]::UTF8, "application/json")

try {
    $response = $client.PostAsync("http://localhost:3000/crawler/start", $content).Result
    Write-Host "Status: $($response.StatusCode)"
    Write-Host "Response: $($response.Content.ReadAsStringAsync().Result)"
} catch {
    Write-Host "Error: $_"
}
