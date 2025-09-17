#!/bin/bash

# K8s 多集群资源监控系统启动脚本

set -e

echo "=========================================="
echo "  K8s 多集群资源监控系统启动脚本"
echo "=========================================="

# 配置文件路径
CONFIG_FILE=${CONFIG_FILE:-"config.toml"}

# 检查配置文件是否存在
if [ ! -f "$CONFIG_FILE" ]; then
    echo "错误: 配置文件不存在: $CONFIG_FILE"
    echo "请先复制 config.toml 文件并配置数据库连接信息"
    exit 1
fi

echo "使用配置文件: $CONFIG_FILE"

# 检查 Go 版本
echo "检查 Go 版本..."
go version

# 下载依赖
echo "下载 Go 模块依赖..."
go mod tidy

# 构建应用
echo "构建应用..."
go build -o bin/cluster-resource-insight cmd/main.go

if [ $? -ne 0 ]; then
    echo "构建失败！请检查错误信息。"
    exit 1
fi

echo "构建成功！"

# 检查是否需要执行数据库迁移
if [ "$1" = "migrate" ]; then
    echo ""
    echo "=========================================="
    echo "  执行数据库迁移"
    echo "=========================================="
    
    ./bin/cluster-resource-insight -config="$CONFIG_FILE" -migrate
    
    echo "数据库迁移完成！"
    exit 0
fi

echo ""
echo "=========================================="
echo "  使用方法"
echo "=========================================="
echo ""
echo "1. 配置数据库连接信息："
echo "   编辑 config.toml 文件中的 [database] 部分"
echo ""
echo "2. 首次运行，执行数据库迁移："
echo "   ./start.sh migrate"
echo ""
echo "3. 启动应用："
echo "   ./start.sh run"
echo ""
echo "4. 使用自定义配置文件："
echo "   CONFIG_FILE=/path/to/your/config.toml ./start.sh run"
echo ""
echo "5. 敏感信息可通过环境变量设置："
echo "   DB_PASSWORD=your_password ./start.sh run"
echo ""
echo "注意事项："
echo "- 确保 MySQL 服务已启动"
echo "- 确保数据库已创建（默认: cluster_resource_insight）"
echo "- 首次运行需要执行数据库迁移"
echo "- 生产环境请修改配置文件中的加密密钥"
echo ""

# 如果没有指定参数，显示使用说明并退出
if [ $# -eq 0 ]; then
    echo "请选择操作："
    echo "  ./start.sh migrate    - 执行数据库迁移"
    echo "  ./start.sh run        - 启动应用"
    exit 0
fi

# 启动应用
if [ "$1" = "run" ]; then
    echo "=========================================="
    echo "  启动应用"
    echo "=========================================="
    echo "配置文件: $CONFIG_FILE"
    echo ""
    echo "启动中..."
    
    ./bin/cluster-resource-insight -config="$CONFIG_FILE"
fi