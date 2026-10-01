# RightHand Go - Global Test Script

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  RightHand Go - Global Test" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Change to script directory
Set-Location $PSScriptRoot

# Step 1: Check Go environment
Write-Host "[Step 1] Checking Go environment..." -ForegroundColor Yellow
$goVersion = go version
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Go is not installed or not in PATH" -ForegroundColor Red
    exit 1
}
Write-Host "$goVersion" -ForegroundColor Green
Write-Host ""

# Step 2: Setup environment
Write-Host "[Step 2] Setting up environment..." -ForegroundColor Yellow
$env:GOPATH = ""
go env -w GOPROXY=https://goproxy.cn,direct
Write-Host "OK - Environment configured" -ForegroundColor Green
Write-Host ""

# Step 3: Tidy dependencies
Write-Host "[Step 3] Tidying dependencies..." -ForegroundColor Yellow
go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Failed to tidy dependencies" -ForegroundColor Red
    exit 1
}
Write-Host "OK - Dependencies tidied" -ForegroundColor Green
Write-Host ""

# Step 4: Run tests
Write-Host "[Step 4] Running tests..." -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Cyan
go test -v ./...

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "  All tests passed!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Red
    Write-Host "  Some tests failed!" -ForegroundColor Red
    Write-Host "========================================" -ForegroundColor Red
}
Write-Host ""
