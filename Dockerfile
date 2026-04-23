# ===================== 构建阶段 =====================
FROM node:22-alpine AS web-builder

WORKDIR /app

COPY web/package*.json ./
RUN npm ci

COPY web/ ./
RUN NODE_ENV=prod npm run build

# ===================== 构建阶段 =====================
FROM golang:1.25 AS builder

WORKDIR /app

# 先复制 go.mod/go.sum 并拉依赖，加快重复构建速度
COPY go.mod go.sum ./
RUN go mod tidy

# 复制项目源码并编译
COPY . .
COPY --from=web-builder /app/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o unimessage .

# ===================== 运行阶段 =====================
FROM alpine:latest

WORKDIR /app

# 仅拷贝编译好的二进制和必要配置
COPY --from=builder /app/unimessage /app/unimessage
COPY conf ./conf

EXPOSE 8081

CMD ["/app/unimessage"]
