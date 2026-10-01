# RightHand Application Deployment Script
# Build frontend and backend + Create Docker images + Update and restart containers

param(
    [switch]$SkipBuild,
    [switch]$SkipFrontend,
    [switch]$SkipBackend
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  RightHand Deployment Script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 1. Build backend
if (-not $SkipBuild -and -not $SkipBackend) {
    Write-Host "[1/4] Building backend..." -ForegroundColor Yellow
    Set-Location "$ProjectRoot\righthand-go"
    
    $env:CGO_ENABLED = 0
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    
    go build -o righthand-server ./cmd/server/main.go
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "      Backend build success!" -ForegroundColor Green
    } else {
        Write-Host "      Backend build failed!" -ForegroundColor Red
        exit 1
    }
    Write-Host ""
}

# 2. Build frontend
if (-not $SkipBuild -and -not $SkipFrontend) {
    Write-Host "[2/4] Building frontend..." -ForegroundColor Yellow
    Set-Location "$ProjectRoot\righthand-frontend"
    
    npm run build
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "      Frontend build success!" -ForegroundColor Green
    } else {
        Write-Host "      Frontend build failed!" -ForegroundColor Red
        exit 1
    }
    Write-Host ""
}

# 3. Stop containers and build images
Write-Host "[3/4] Stopping old containers and building new images..." -ForegroundColor Yellow
Set-Location $ProjectRoot

# 停止并删除所有相关容器（包括 docker-compose 和手动创建的）
docker-compose down

# 强制删除可能存在的同名容器（手动创建的容器不在 docker-compose 管理范围内）
$containers = @("righthand-go", "righthand-frontend")
foreach ($container in $containers) {
    if (docker ps -aq --filter "name=^/$container$") {
        Write-Host "      Removing existing container: $container" -ForegroundColor White
        docker rm -f $container | Out-Null
    }
}

# 安全删除镜像，忽略不存在的镜像
$images = @("righthand-go", "righthand-righthand-go", "righthand-righthand-frontend")
foreach ($img in $images) {
    if (docker images -q $img) {
        docker rmi -f $img | Out-Null
    }
}

docker build --no-cache -t righthand-go -f righthand-go/Dockerfile .

if ($LASTEXITCODE -eq 0) {
    Write-Host "      Image build success!" -ForegroundColor Green
} else {
    Write-Host "      Image build failed!" -ForegroundColor Red
    exit 1
}
Write-Host ""

# 4. Start containers
Write-Host "[4/4] Starting Docker containers..." -ForegroundColor Yellow
docker-compose up -d

if ($LASTEXITCODE -eq 0) {
    Write-Host "      Containers started!" -ForegroundColor Green
} else {
    Write-Host "      Container start failed!" -ForegroundColor Red
    exit 1
}
Write-Host ""

# 5. Show status
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Deployment Complete - Service Status" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

docker ps --filter "name=righthand" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

Write-Host ""
Write-Host "Access URLs:" -ForegroundColor White
Write-Host "  Local:      http://localhost" -ForegroundColor Green
Write-Host "  LAN:        http://10.252.20.59" -ForegroundColor Green
Write-Host ""
Write-Host "Backend API: http://localhost:8090/api/" -ForegroundColor Green
Write-Host ""