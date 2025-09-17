# K8s 集群资源监控系统

## 项目简介

这是一个用于监控和分析 Kubernetes 集群资源配置的工具，帮助识别资源配置不合理的 Pod，优化集群资源利用率。

## 功能特性

- 🔍 **全面扫描**: 扫描所有命名空间下的 Pod 资源配置
- 📊 **智能分析**: 识别内存和 CPU 利用率过低的 Pod
- 🎯 **问题定位**: 自动标记配置不合理的资源
- 📈 **可视化展示**: Web 界面展示分析结果
- 📤 **数据导出**: 支持导出 CSV 格式报告

## 系统要求

- Go 1.24+
- Kubernetes 集群访问权限
- Metrics Server (用于获取实际使用情况)

## 快速开始

### 1. 构建项目

```bash
# 克隆代码后进入项目目录
cd ClusterResourceInsight

# 构建项目
./build.sh
```

### 2. 运行应用

```bash
# 使用默认 kubeconfig
./bin/cluster-resource-insight

# 或指定 kubeconfig 路径
./bin/cluster-resource-insight -kubeconfig=/path/to/your/kubeconfig
```

### 3. 访问 Web 界面

打开浏览器访问: http://localhost:8080

## 判断标准

### 资源配置不合理标准

1. **内存利用率过低**
   - 内存使用量/内存请求 < 20%
   - 内存使用量/内存限制 < 15%

2. **CPU 利用率过低**
   - CPU 使用量/CPU 请求 < 15%
   - CPU 使用量/CPU 限制 < 10%

3. **配置缺失**
   - 未设置 Request 或 Limit
   - Request 和 Limit 差异过大 (超过 300%)

## API 接口

### 获取分析结果
```
GET /api/v1/analysis
```

### 获取 Pod 数据
```
GET /api/v1/pods?limit=50&only_problems=true
```

### 健康检查
```
GET /api/v1/health
```

## 数据格式示例

```json
{
  "total_pods": 156,
  "unreasonable_pods": 23,
  "top50_problems": [
    {
      "pod_name": "ds-2048-test-5l7fz",
      "namespace": "default",
      "node_name": "worker-1",
      "memory_usage": 1572864,
      "memory_request": 67108864,
      "memory_req_pct": 2.44,
      "memory_limit": 67108864,
      "memory_limit_pct": 2.44,
      "status": "不合理",
      "issues": ["内存请求利用率过低", "内存限制利用率过低"]
    }
  ],
  "generated_at": "2024-01-15T10:30:00Z"
}
```

## 项目结构

```
ClusterResourceInsight/
├── cmd/                    # 应用入口
│   └── main.go
├── internal/               # 内部代码
│   ├── api/               # API 处理器
│   ├── collector/         # 数据收集器
│   └── config/           # 配置管理
├── web/                   # 前端资源
│   ├── templates/        # HTML 模板
│   └── static/          # 静态文件
├── build.sh              # 构建脚本
├── go.mod               # Go 模块定义
└── README.md           # 项目说明
```

## 注意事项

1. **权限要求**: 需要有读取 pods、namespaces 和 metrics 的权限
2. **Metrics Server**: 必须安装 metrics-server 才能获取实际资源使用情况
3. **网络访问**: 确保能够访问 Kubernetes API Server

## 故障排除

### 常见问题

1. **无法获取 Metrics 数据**
   - 确认 metrics-server 已安装并运行
   - 检查 RBAC 权限配置

2. **连接集群失败**
   - 验证 kubeconfig 文件路径和内容
   - 确认集群网络连通性

3. **构建失败**
   - 检查 Go 版本是否符合要求
   - 确认网络能够下载依赖包

## 开发说明

项目使用 Go 开发，主要依赖：
- `k8s.io/client-go` - Kubernetes 客户端
- `k8s.io/metrics` - Metrics API 客户端
- `gin-gonic/gin` - Web 框架

## 许可证

MIT License