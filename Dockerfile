# 阶段 1: 编译 Go 项目
FROM golang:1.25-alpine AS builder

# 配置 Alpine 国内镜像源（加速 apk add）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache git

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 配置 Go 模块国内代理（加速 go mod download）
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译 Go 项目
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/server cmd/server/main.go

# 编译健康检查工具
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/healthcheck healthcheck.go

# 阶段 2: 创建最小化运行镜像
FROM scratch

# 复制 CA 证书用于 HTTPS 请求
COPY ca-certificates.crt /etc/ssl/certs/

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/server /app/server
COPY --from=builder /app/healthcheck /app/healthcheck
COPY application.properties /app/application.properties

# 设置工作目录
WORKDIR /app

# 设置环境变量
ENV APP_ENV=production

# 暴露端口
EXPOSE 28080

# 启动应用
ENTRYPOINT ["/app/server"]
