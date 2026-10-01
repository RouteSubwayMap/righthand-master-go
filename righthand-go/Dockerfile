# 先在宿主机编译
# CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o righthand-server ./cmd/server/main.go

# 生产阶段
FROM alpine:latest

WORKDIR /app

# 设置时区为东八区
RUN apk add --no-cache tzdata
ENV TZ=Asia/Shanghai
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

COPY righthand-go/righthand-server .
COPY righthand-go/configs ./configs

EXPOSE 8090

CMD ["./righthand-server"]