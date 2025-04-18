#第一阶段:使用 golang:1.23 镜像进行构建FRoM golang:1.23 As builder
FROM golang:1.23 AS builder

#复制项目文件到工作目录
COPY . /app

#设置工作目录
WORKDIR /app

ARG TARGET="gw"

#下载并整理 Go 模块依赖，构建Go 应用程序
RUN CGO_ENABLED=0
RUN GOOS=linux
RUN GOPROXY=https://goproxy.cn,direct make ${TARGET} 

#第二阶段:使用 debian:stable-slim 作为基础镜像
FROM debian:stable-slim

#安装必要的 CA 证书，以便应用程序可以进行 HTTPS 请求
RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

# 设置时区为 Asia/Shanghai
ENV TZ=Asia/Shanghai

# 安装 tzdata 包（如果需要）
RUN apt-get update && apt-get install -y tzdata

# 定义构建参数 CONFIGS，默认值为 "prod"
ARG CONFIGS="prod"
ARG TARGET="gw"

# 无论 CONFIGS 为何值，都要执行的复制操作
COPY --from=builder /app/bin /app
COPY --from=builder /app/configs /app/configs
COPY --from=builder /app/configs-dev /app/configs-dev
COPY --from=builder /app/configs-cron /app/configs-cron

RUN if [ "$CONFIGS" = "test" ]; then \
        cp /app/configs-dev/config.yaml /app/configs/config.yaml; \
        cp /app/configs-dev/msg-local.yaml /app/configs/msg-local.yaml; \
        rm -rf /app/configs-dev; \
    else \
        rm -rf /app/configs-dev; \
    fi

RUN if [ "${TARGET}" = "job" ]; then \
        cp /app/configs-cron/config.yaml /app/configs/config.yaml; \
        rm -rf /app/configs-cron; \
    else \
        rm -rf /app/configs-cron; \
    fi

#设置工作目录
WORKDIR /app

#暴露应用程序监听的端口
EXPOSE 9000
EXPOSE 9001

#设置容器启动时执行的命令
CMD ["/app/server"]
