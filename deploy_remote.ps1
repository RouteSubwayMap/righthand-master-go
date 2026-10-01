param(
    [string]$ServerIP = "10.252.27.13",
    
    [string]$SSHUser = "appuser",
    
    [string]$SSHPort = "22",
    [string]$RemotePath = "",
    [string]$DBHost = "127.0.0.1",
    [string]$DBPort = "3306",
    [string]$DBName = "righthand",
    [string]$DBUser = "root",
    [string]$DBPassword = "root",
    
    [switch]$SkipBuild,
    [switch]$SkipPackage
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent $MyInvocation.MyCommand.Path

if ([string]::IsNullOrEmpty($RemotePath)) {
    $RemotePath = "/home/$SSHUser/righthand"
}

$SSHHost = "$SSHUser@$ServerIP"

function Write-Cyan {
    Write-Host $args[0] -ForegroundColor Cyan
}

function Write-Yellow {
    Write-Host $args[0] -ForegroundColor Yellow
}

function Write-Green {
    Write-Host $args[0] -ForegroundColor Green
}

function Write-Red {
    Write-Host $args[0] -ForegroundColor Red
}

Write-Cyan "========================================"
Write-Cyan "  RightHand Remote Deployment Script"
Write-Cyan "========================================"
Write-Host ""

Write-Yellow "[1/5] Checking SSH connectivity..."
Write-Host "      Please enter password when prompted..." -ForegroundColor White
try {
    & ssh -o StrictHostKeyChecking=no -p $SSHPort "$SSHHost" "echo 'SSH connection OK'"
    if ($LASTEXITCODE -eq 0) {
        Write-Green "      SSH connection OK!"
    } else {
        Write-Red "      SSH connection failed!"
        exit 1
    }
} catch {
    Write-Red "      SSH connection error: $_"
    exit 1
}
Write-Host ""

if (-not $SkipBuild) {
    Write-Yellow "[2/5] Building locally (cross-compile for Linux)..."
    
    # 交叉编译 Go 后端
    Write-Host "      Building backend..." -ForegroundColor White
    $env:CGO_ENABLED = 0
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    
    Push-Location "$ProjectRoot\righthand-go"
    go build -o righthand-server ./cmd/server/main.go
    if ($LASTEXITCODE -ne 0) {
        Write-Red "      Backend build failed!"
        Pop-Location
        exit 1
    }
    Pop-Location
    Write-Green "      Backend build success!"
    
    # 构建前端
    Write-Host "      Building frontend..." -ForegroundColor White
    Push-Location "$ProjectRoot\righthand-frontend"
    npm run build
    if ($LASTEXITCODE -ne 0) {
        Write-Red "      Frontend build failed!"
        Pop-Location
        exit 1
    }
    Pop-Location
    Write-Green "      Frontend build success!"
    Write-Host ""
}

$stepNum = if ($SkipBuild) { "2/4" } else { "3/5" }
Write-Yellow "[$stepNum] Preparing package..."
$packageFile = "$ProjectRoot\righthand-deploy.tar.gz"

if ($SkipPackage) {
    if (-not (Test-Path $packageFile)) {
        Write-Red "      SkipPackage specified but $packageFile not found!"
        exit 1
    }
    Write-Green "      Using existing package file (--skip-package)"
} else {
    if (Test-Path $packageFile) {
        Remove-Item $packageFile -Force
    }
    try {
        & tar -czf "$packageFile" --exclude="node_modules" --exclude=".git" --exclude="*.tar.gz" --exclude="*.zip" -C "$ProjectRoot" .
        if ($LASTEXITCODE -ne 0) {
            throw "tar command failed"
        }
        Write-Green "      Project files packaged successfully!"
    } catch {
        Write-Red "      Failed to package project files: $_"
        exit 1
    }
}
Write-Host ""

$stepNum2 = if ($SkipBuild) { "3/4" } else { "4/5" }
Write-Yellow "[$stepNum2] Uploading and deploying to remote server..."
Write-Host "      Please enter password when prompted (multiple times)..." -ForegroundColor White
try {
    # 先在远程创建目录
    & ssh -o StrictHostKeyChecking=no -p $SSHPort "$SSHHost" "mkdir -p $RemotePath"
    if ($LASTEXITCODE -ne 0) {
        Write-Red "      Failed to create remote directory!"
        exit 1
    }

    # 远程部署脚本始终跳过编译（已在本地完成）
    $remoteCmd = "cd $RemotePath; rm -rf righthand; mkdir -p righthand; tar -xzf righthand-deploy.tar.gz -C righthand; cd righthand; chmod +x deploy.sh; DATABASE_HOST=$DBHost DATABASE_PORT=$DBPort DATABASE_USERNAME=$DBUser DATABASE_PASSWORD=$DBPassword DATABASE_DBNAME=$DBName ./deploy.sh --skip-build"

    Write-Host "      Uploading package file..."
    & scp -o StrictHostKeyChecking=no -P $SSHPort "$packageFile" "$SSHHost`:$RemotePath/"
    if ($LASTEXITCODE -ne 0) {
        Write-Red "      Failed to upload project files!"
        exit 1
    }

    Write-Host "      Executing remote deployment..."
    & ssh -o StrictHostKeyChecking=no -p $SSHPort "$SSHHost" $remoteCmd
    if ($LASTEXITCODE -eq 0) {
        Write-Green "      Remote deployment completed successfully!"
    } else {
        Write-Red "      Remote deployment failed!"
        exit 1
    }
} catch {
    Write-Red "      Deployment error: $_"
    exit 1
}
Write-Host ""

Write-Cyan "========================================"
Write-Cyan "  Remote Deployment Complete"
Write-Cyan "========================================"
Write-Host ""
Write-Cyan "Access URLs:"
Write-Green "  Local:      http://$ServerIP"
Write-Green "  Backend API: http://$ServerIP:8090/api/"
Write-Host ""