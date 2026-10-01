#!/bin/bash

set -e

PROJECT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)

print_cyan() {
    echo -e "\033[36m$1\033[0m"
}

print_yellow() {
    echo -e "\033[33m$1\033[0m"
}

print_green() {
    echo -e "\033[32m$1\033[0m"
}

print_red() {
    echo -e "\033[31m$1\033[0m"
}

print_cyan "========================================"
print_cyan "  RightHand Deployment Script (Ubuntu)"
print_cyan "========================================"
echo ""

SKIP_BUILD=false
SKIP_FRONTEND=false
SKIP_BACKEND=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        --skip-frontend)
            SKIP_FRONTEND=true
            shift
            ;;
        --skip-backend)
            SKIP_BACKEND=true
            shift
            ;;
        *)
            echo "未知参数: $1"
            exit 1
            ;;
    esac
done

if [[ "$SKIP_BUILD" == false && "$SKIP_BACKEND" == false ]]; then
    print_yellow "[1/4] Building backend..."
    cd "$PROJECT_ROOT/righthand-go"
    
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o righthand-server ./cmd/server/main.go
    
    if [[ $? -eq 0 ]]; then
        print_green "      Backend build success!"
    else
        print_red "      Backend build failed!"
        exit 1
    fi
    echo ""
fi

if [[ "$SKIP_BUILD" == false && "$SKIP_FRONTEND" == false ]]; then
    print_yellow "[2/4] Building frontend..."
    cd "$PROJECT_ROOT/righthand-frontend"
    
    npm run build
    
    if [[ $? -eq 0 ]]; then
        print_green "      Frontend build success!"
    else
        print_red "      Frontend build failed!"
        exit 1
    fi
    echo ""
fi

print_yellow "[3/4] Stopping old containers and building new images..."
cd "$PROJECT_ROOT"

docker compose -f docker-compose.ubuntu.yml down 2>/dev/null || true

containers=("righthand-go" "righthand-frontend")
for container in "${containers[@]}"; do
    if docker ps -aq --filter "name=^/$container$" | grep -q .; then
        echo "      Removing existing container: $container"
        docker rm -f "$container" >/dev/null 2>&1
    fi
done

images=("righthand-go" "righthand-righthand-go" "righthand-righthand-frontend")
for img in "${images[@]}"; do
    if docker images -q "$img" | grep -q .; then
        docker rmi -f "$img" >/dev/null 2>&1
    fi
done

docker build --no-cache -t righthand-go -f righthand-go/Dockerfile .

if [[ $? -eq 0 ]]; then
    print_green "      Image build success!"
else
    print_red "      Image build failed!"
    exit 1
fi
echo ""

print_yellow "[4/4] Starting Docker containers..."
DATABASE_HOST=${DATABASE_HOST:-172.17.0.1} \
DATABASE_PORT=${DATABASE_PORT:-3306} \
DATABASE_USERNAME=${DATABASE_USERNAME:-root} \
DATABASE_PASSWORD=${DATABASE_PASSWORD:-password} \
DATABASE_DBNAME=${DATABASE_DBNAME:-righthand_db} \
DATABASE_LOC=${DATABASE_LOC:-Asia/Shanghai} \
docker compose -f docker-compose.ubuntu.yml up -d

if [[ $? -eq 0 ]]; then
    print_green "      Containers started!"
else
    print_red "      Container start failed!"
    exit 1
fi
echo ""

print_cyan "========================================"
print_cyan "  Deployment Complete - Service Status"
print_cyan "========================================"
echo ""

docker ps --filter "name=righthand" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo ""
print_cyan "Access URLs:"
print_green "  Local:      http://localhost"
HOST_IP=$(hostname -I | awk '{print $1}')
print_green "  LAN:        http://$HOST_IP"
echo ""
print_green "Backend API: http://localhost:8090/api/"
echo ""