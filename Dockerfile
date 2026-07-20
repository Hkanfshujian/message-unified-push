# ===================== 前端构建阶段 =====================
ARG NODE_IMAGE=node:22-alpine
ARG GO_IMAGE=golang:1.25
ARG RUNTIME_IMAGE=alpine:latest

FROM ${NODE_IMAGE} AS web-builder

WORKDIR /app

COPY web/package*.json ./
COPY web/scripts ./scripts
RUN npm ci

COPY web/ ./
RUN NODE_ENV=prod npm run build

# ===================== 后端构建阶段 =====================
FROM ${GO_IMAGE} AS builder

ARG GOPROXY=https://proxy.golang.org,direct
ENV GOPROXY=${GOPROXY}

WORKDIR /app

# 先复制 go.mod/go.sum 并下载依赖，加快重复构建速度
COPY go.mod go.sum ./
RUN go mod download

# 复制项目源码并编译
COPY . .
COPY --from=web-builder /app/dist ./web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o unimessage .

# ===================== 运行阶段 =====================
FROM ${RUNTIME_IMAGE}

WORKDIR /app

RUN apk add --no-cache curl ca-certificates

# 仅拷贝编译好的二进制和默认配置，不将本地 conf/app.ini 打进镜像
COPY --from=builder /app/unimessage /app/unimessage
COPY conf/app.example.ini ./conf/app.example.ini

EXPOSE 8081

CMD ["/app/unimessage"]
