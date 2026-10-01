# ============================================================
# RightHand Release Build Script (PowerShell - Windows)
# Cross-compile Go backend + build frontend + package for Linux deployment
# Usage: .\scripts\build-release.ps1
# ============================================================
$ErrorActionPreference = "Stop"

$ROOT = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)

# Version: read from VERSION file, fallback to git commit hash
$RELEASE_VER = if (Test-Path "$ROOT/VERSION") { (Get-Content "$ROOT/VERSION").Trim() } else { "dev" }
$GIT_COMMIT = (git -C $ROOT rev-parse --short HEAD 2>$null) -replace "`n|`r", ""
if (-not $GIT_COMMIT) { $GIT_COMMIT = "unknown" }
$BUILD_DATE = Get-Date -Format "yyyyMMdd"
$BUILD_TIME = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")

Write-Host "=== RightHand Release Build v${RELEASE_VER} (${BUILD_DATE}) ===" -ForegroundColor Green
Write-Host "  Git Commit: $GIT_COMMIT"
Write-Host "  Build Time: $BUILD_TIME"
Write-Host ""

$STAGE_NAME = "righthand-${BUILD_DATE}-v${RELEASE_VER}"
$DIST = Join-Path $ROOT "dist"
$STAGE = Join-Path $DIST $STAGE_NAME

# Clean old dist directory
if (Test-Path $DIST) {
    Remove-Item -Recurse -Force $DIST
    Write-Host "-> cleaned old dist directory" -ForegroundColor Cyan
}

# Create staging directory structure
New-Item -ItemType Directory -Force -Path "$STAGE/bin", "$STAGE/web", "$STAGE/configs", "$STAGE/sql", "$STAGE/nginx", "$STAGE/systemd" | Out-Null
Write-Host "-> created staging directory structure" -ForegroundColor Cyan

# ============================================================
# 1. Build Go backend for Linux amd64 (static binary)
# ============================================================
Write-Host ""
Write-Host "[1/3] Building Go backend for Linux amd64..." -ForegroundColor Yellow

$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"

# Inject version info via ldflags
$LDFLAGS = "-s -w -X main.Version=$RELEASE_VER -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$GIT_COMMIT"

Push-Location "$ROOT/righthand-go"
go build -trimpath -ldflags $LDFLAGS -o "$STAGE/bin/righthand-server" "./cmd/server/main.go"
Pop-Location

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Backend build failed!" -ForegroundColor Red
    exit 1
}
Write-Host "      Backend build success -> bin/righthand-server" -ForegroundColor Green

# ============================================================
# 2. Build frontend
# ============================================================
Write-Host ""
Write-Host "[2/3] Building frontend..." -ForegroundColor Yellow

Push-Location "$ROOT/righthand-frontend"
npm install
npm run build
Pop-Location

if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Frontend build failed!" -ForegroundColor Red
    exit 1
}

# Copy frontend build output
Copy-Item -Recurse "$ROOT/righthand-frontend/dist/*" "$STAGE/web/" -Force
Write-Host "      Frontend build success -> web/" -ForegroundColor Green

# ============================================================
# 3. Collect deployment resources
# ============================================================
Write-Host ""
Write-Host "[3/3] Collecting deployment resources..." -ForegroundColor Yellow

# Backend config template
Copy-Item "$ROOT/scripts/config.yaml.tpl" "$STAGE/configs/config.yaml" -Force
Write-Host "      configs/config.yaml"

# SQL init script
Copy-Item "$ROOT/righthand-go/sql/init-database.sql" "$STAGE/sql/" -Force
Write-Host "      sql/init-database.sql"

# Nginx config
Copy-Item "$ROOT/scripts/righthand.conf" "$STAGE/nginx/" -Force
Write-Host "      nginx/righthand.conf"

# systemd service unit
Copy-Item "$ROOT/scripts/righthand.service" "$STAGE/systemd/" -Force
Write-Host "      systemd/righthand.service"

# Deployment scripts
Copy-Item "$ROOT/scripts/install.sh" "$STAGE/install.sh" -Force
Copy-Item "$ROOT/scripts/righthand-start.sh" "$STAGE/righthand-start.sh" -Force
Copy-Item "$ROOT/scripts/righthand-stop.sh" "$STAGE/righthand-stop.sh" -Force
Write-Host "      install.sh / righthand-start.sh / righthand-stop.sh"

# ============================================================
# 4. Fix line endings: CRLF -> LF (Linux compatible)
# ============================================================
Write-Host ""
Write-Host "-> fixing line endings (CRLF -> LF)..." -ForegroundColor Cyan
Get-ChildItem -Path $STAGE -Include "*.sh", "*.conf", "*.service", "*.yaml", "*.sql" -Recurse | ForEach-Object {
    $content = [IO.File]::ReadAllText($_.FullName) -replace "`r`n", "`n"
    [IO.File]::WriteAllText($_.FullName, $content)
}
Write-Host "      line endings fixed"

# ============================================================
# 5. Package
# ============================================================
Write-Host ""
Write-Host "-> packaging..." -ForegroundColor Cyan
Push-Location $DIST
tar czf "$STAGE_NAME.tar.gz" "$STAGE_NAME"
Pop-Location
Write-Host "      packaged: $STAGE_NAME.tar.gz"
Write-Host "      note: run 'chmod +x install.sh' after extraction on Linux" -ForegroundColor DarkGray

# ============================================================
# Summary
# ============================================================
Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "  Build Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "  Package: $DIST\$STAGE_NAME.tar.gz" -ForegroundColor White
Write-Host "  Staging: $STAGE" -ForegroundColor White
Write-Host ""
Write-Host "  On target server:" -ForegroundColor Cyan
Write-Host "    tar xzf $STAGE_NAME.tar.gz && cd $STAGE_NAME" -ForegroundColor White
Write-Host "    chmod +x *.sh bin/righthand-server" -ForegroundColor White
Write-Host "    vim configs/config.yaml    # edit DB password and JWT secret" -ForegroundColor White
Write-Host "    vim install.sh             # set AUTO_INIT_DB=true, DB_PASS etc." -ForegroundColor White
Write-Host "    sudo ./install.sh          # one-click deploy" -ForegroundColor White
Write-Host ""
