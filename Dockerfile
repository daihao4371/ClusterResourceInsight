# 使用轻量级 Alpine 镜像作为基础镜像
FROM alpine:3.19

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata curl

# 设置时区
ENV TZ=Asia/Shanghai

# 创建应用用户和目录
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup && \
    mkdir -p /app/web/dist /app/web-legacy /app/logs && \
    chown -R appuser:appgroup /app

# 设置工作目录
WORKDIR /app

# 复制后端二进制文件（需要先编译：go build -o cluster-resource-insight ./cmd/main.go）
COPY cluster-resource-insight ./

# 复制配置文件
COPY config.toml ./
COPY init.sql ./

# 复制前端构建文件（需要先构建：cd web && npm run build）
COPY web/dist/ ./web/dist/
COPY web/dist/ ./web/

# 复制传统web文件作为备份
COPY web-legacy/ ./web-legacy/

# 设置文件权限
RUN chmod +x ./cluster-resource-insight && \
    chown -R appuser:appgroup /app

# 创建健康检查脚本
RUN echo '#!/bin/sh\ncurl -f http://localhost:9999/api/health || exit 1' > /app/health-check.sh && \
    chmod +x /app/health-check.sh

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 9999

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD /app/health-check.sh

# 启动应用
CMD ["./cluster-resource-insight"]