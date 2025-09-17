#!/bin/bash

# K8s 集群资源监控启动脚本

echo "开始构建 K8s 集群资源监控系统..."

# 检查 Go 版本
go version

# 下载依赖
echo "下载 Go 模块依赖..."
go mod tidy

# 构建应用
echo "构建应用..."
go build -o bin/cluster-resource-insight cmd/main.go

if [ $? -eq 0 ]; then
    echo "构建成功！"
    echo ""
    echo "使用方法："
    echo "1. 确保 kubectl 配置正确："
    echo "   kubectl config current-context"
    echo ""
    echo "2. 启动应用："
    echo "   ./bin/cluster-resource-insight"
    echo ""
    echo "3. 或者指定 kubeconfig 路径："
    echo "   ./bin/cluster-resource-insight -kubeconfig=/path/to/your/kubeconfig"
    echo ""
    echo "4. 访问 Web 界面："
    echo "   http://localhost:8080"
    echo ""
    echo "注意："
    echo "- 需要安装 metrics-server 才能获取实际使用情况"
    echo "- 确保有足够的 RBAC 权限读取 pods 和 metrics"
else
    echo "构建失败！请检查错误信息。"
    exit 1
fi