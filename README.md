# ClusterResourceInsight - K8s 多集群资源监控平台

## 项目简介

ClusterResourceInsight 是一个专业的 Kubernetes 多集群资源监控与分析平台，提供实时的集群状态监控、资源配置分析、智能告警系统和直观的可视化界面。帮助运维团队优化集群资源利用率，快速定位和解决资源配置问题。

## ✨ 核心功能

### 🌐 多集群管理
- **集群配置管理**: 支持多种认证方式 (kubeconfig/token/cert)
- **实时连接监控**: 自动检测集群连接状态和健康度
- **批量操作**: 支持批量测试和管理多个集群

### 📊 智能监控分析
- **资源配置分析**: 深度分析 Pod 内存/CPU 配置合理性
- **实时数据收集**: 基于 Metrics Server 获取真实资源使用情况
- **问题自动识别**: 智能标记配置不合理的 Pod 和资源浪费

### 🚨 告警与活动系统
- **实时告警**: 基于真实集群状态生成智能告警
- **活动追踪**: 记录集群连接、数据收集、状态变更等系统活动
- **告警操作**: 支持告警解决、忽略等操作管理

### 📈 可视化仪表板
- **系统总览**: 实时显示集群状态、Pod 统计、资源效率等关键指标
- **趋势分析**: 提供资源使用趋势图表和历史数据分析
- **分页查询**: 支持高效的分页查询和数据筛选

### 🔧 高级功能
- **调度管理**: 支持定时数据收集和任务调度
- **历史数据**: 完整的历史数据查询和统计分析
- **数据导出**: 支持 CSV 等格式的数据导出

## 系统要求

- Go 1.19+
- Node.js 18+ (前端开发)
- MySQL 5.7+ 或 8.0+
- Kubernetes 集群访问权限
- Metrics Server (推荐，用于获取实际使用情况)

## 🚀 快速开始

### 1. 环境准备

```bash
# 克隆项目
git clone https://github.com/your-org/ClusterResourceInsight.git
cd ClusterResourceInsight

# 安装前端依赖
cd web
npm install
npm run build
cd ..
```

### 2. 配置数据库

```bash
# 创建数据库
mysql -u root -p
CREATE DATABASE cluster_resource_insight CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 复制配置文件并修改数据库连接信息
cp config.toml.example config.toml
# 编辑 config.toml 中的数据库配置
```

### 3. 运行应用

```bash
# 构建并运行
go run cmd/main.go

# 或构建后运行
go build -o bin/cluster-resource-insight cmd/main.go
./bin/cluster-resource-insight
```

### 4. 访问系统

- **Web界面**: http://localhost:9999
- **API文档**: http://localhost:9999/api/v1/health
- **集群管理**: http://localhost:9999/clusters

## 📖 使用指南

### 集群配置

1. 访问 Web 界面，进入集群管理页面
2. 点击"添加集群"，填写集群信息：
   - 集群名称和别名
   - API Server 地址
   - 认证方式 (推荐使用 kubeconfig)
   - 采集间隔设置

### 监控分析

1. **系统总览**: 查看集群状态、Pod 统计、资源效率等关键指标
2. **资源分析**: 查看问题 Pod 列表，支持按集群筛选和分页
3. **实时活动**: 监控系统活动和告警信息
4. **趋势分析**: 查看历史数据和资源使用趋势

### 告警管理

- **查看告警**: 在系统总览页面查看最新告警
- **处理告警**: 点击 ✓ 标记告警为已解决，点击 ✗ 忽略告警
- **告警详情**: 点击告警可查看详细信息

## 📋 判断标准

### 资源配置不合理标准

系统采用多维度分析标准，自动识别资源配置问题：

1. **内存利用率过低**
   - 内存使用量/内存请求 < 20%
   - 内存使用量/内存限制 < 15%

2. **CPU 利用率过低**
   - CPU 使用量/CPU 请求 < 15%
   - CPU 使用量/CPU 限制 < 10%

3. **配置缺失**
   - 未设置 Request 或 Limit
   - Request 和 Limit 差异过大 (超过 300%)

4. **资源浪费程度**
   - **严重浪费**: 利用率 < 5%
   - **中等浪费**: 利用率 5-15%
   - **轻度浪费**: 利用率 15-20%

## 🔌 API 接口

### 系统管理
```http
# 系统状态
GET /api/v1/health
GET /api/v1/stats

# 集群管理
GET    /api/v1/clusters
POST   /api/v1/clusters
PUT    /api/v1/clusters/{id}
DELETE /api/v1/clusters/{id}
POST   /api/v1/clusters/{id}/test
```

### 资源分析
```http
# 资源分析
GET /api/v1/analysis?page=1&size=50&cluster_name=xxx

# Pod管理
GET /api/v1/pods/search?namespace=xxx&pod_name=xxx
GET /api/v1/pods/problems?page=1&size=20

# 统计信息
GET /api/v1/statistics/top-memory-request
GET /api/v1/statistics/top-cpu-request
GET /api/v1/statistics/namespace-summary
```

### 活动与告警
```http
# 活动管理
GET    /api/v1/activities/recent?limit=10
DELETE /api/v1/activities/cleanup

# 告警管理
GET /api/v1/alerts/recent?limit=10
PUT /api/v1/alerts/{id}/resolve
PUT /api/v1/alerts/{id}/dismiss
GET /api/v1/alerts/{id}
```

### 历史数据
```http
# 趋势数据
GET /api/v1/history/trends?cluster_id=1&hours=24
GET /api/v1/history/system-trends?hours=24

# 数据管理
POST   /api/v1/history/collect
DELETE /api/v1/history/cleanup?retention_days=30
```

## 📄 数据格式示例

### 系统统计响应
```json
{
  "code": 0,
  "data": {
    "total_clusters": 2,
    "online_clusters": 2,
    "total_pods": 53,
    "problem_pods": 23,
    "resource_efficiency": 82.5,
    "cluster_status_distribution": [
      {"name": "在线", "value": 2, "color": "#22c55e"}
    ],
    "last_update": "2024-01-15T10:30:00Z"
  },
  "msg": "操作成功"
}
```

### 告警列表响应
```json
{
  "code": 0,
  "data": {
    "count": 5,
    "data": [
      {
        "id": 123,
        "level": "high",
        "title": "严重资源配置问题",
        "description": "Pod xxx/yyy 资源利用率极低：内存 0.4%, CPU 0.5%",
        "time": "2分钟前",
        "status": "active"
      }
    ]
  }
}
```

### Pod 分析结果
```json
{
  "total_pods": 156,
  "unreasonable_pods": 23,
  "top50_problems": [
    {
      "pod_name": "app-deployment-5l7fz",
      "namespace": "default",
      "cluster_name": "prod-cluster",
      "node_name": "worker-1",
      "memory_usage": 1572864,
      "memory_request": 67108864,
      "memory_req_pct": 2.44,
      "cpu_usage": 2,
      "cpu_request": 100,
      "cpu_req_pct": 2.0,
      "status": "unreasonable",
      "issues": ["内存请求利用率过低", "CPU请求利用率过低"],
      "waste_score": 85.5
    }
  ],
  "generated_at": "2024-01-15T10:30:00Z"
}
```

## 🏗️ 项目架构

```
ClusterResourceInsight/
├── cmd/                    # 应用入口
│   └── main.go            # 主程序入口
├── internal/              # 内部核心代码
│   ├── api/              # API 路由和处理器
│   │   └── handlers.go   # HTTP 处理器
│   ├── collector/        # 数据收集器
│   │   ├── resource_collector.go      # 资源数据收集
│   │   └── multi_cluster_collector.go # 多集群数据收集
│   ├── service/          # 业务逻辑层
│   │   ├── cluster_service.go        # 集群管理服务
│   │   ├── activity_service.go       # 活动和告警服务
│   │   ├── history_service.go        # 历史数据服务
│   │   └── schedule_service.go       # 调度管理服务
│   ├── models/           # 数据模型
│   │   └── models.go     # 数据库模型定义
│   ├── database/         # 数据库管理
│   │   └── database.go   # 数据库连接和初始化
│   ├── config/           # 配置管理
│   │   └── config.go     # 配置文件解析
│   ├── logger/           # 日志系统
│   │   └── logger.go     # 日志配置
│   └── response/         # 响应工具
│       └── response.go   # 统一响应格式
├── web/                   # 前端代码 (Vue.js)
│   ├── src/
│   │   ├── components/   # Vue 组件
│   │   ├── views/        # 页面视图
│   │   ├── stores/       # 状态管理 (Pinia)
│   │   ├── types/        # TypeScript 类型定义
│   │   └── utils/        # 工具函数
│   ├── public/           # 静态资源
│   └── dist/             # 构建输出
├── pkg/                   # 可复用的包
│   ├── statistics/       # 统计工具
│   └── utils/            # 通用工具
├── config.toml           # 配置文件
├── go.mod               # Go 模块定义
├── go.sum               # 依赖校验
└── README.md           # 项目文档
```

## 🛠️ 技术栈

### 后端
- **语言**: Go 1.19+
- **Web框架**: Gin (HTTP服务)
- **数据库**: MySQL + GORM (ORM)
- **K8s客户端**: client-go + metrics API
- **配置管理**: Viper (TOML配置)
- **日志**: 自定义日志系统

### 前端
- **框架**: Vue 3 + TypeScript
- **构建工具**: Vite
- **状态管理**: Pinia
- **UI组件**: 自定义组件 + Tailwind CSS
- **图表**: 自定义图表组件
- **HTTP客户端**: Axios

### 数据库设计
- **cluster_configs**: 集群配置信息
- **pod_metrics_history**: Pod 监控历史数据
- **system_activities**: 系统活动记录
- **alert_history**: 告警历史记录
- **alert_rules**: 告警规则配置
- **system_settings**: 系统配置

## ⚠️ 注意事项

### 权限要求
1. **Kubernetes RBAC**: 需要以下权限
   ```yaml
   rules:
   - apiGroups: [""]
     resources: ["pods", "namespaces", "nodes"]
     verbs: ["get", "list"]
   - apiGroups: ["metrics.k8s.io"]
     resources: ["pods", "nodes"]
     verbs: ["get", "list"]
   ```

2. **数据库权限**: 需要创建表、插入、查询、更新权限

### 部署建议
1. **生产环境**: 建议使用容器化部署 (Docker/Kubernetes)
2. **资源要求**: 最少 2GB 内存，推荐 4GB+
3. **存储**: 数据库建议使用 SSD 存储以提高查询性能
4. **网络**: 确保能访问所有需要监控的 K8s 集群

### 性能优化
1. **缓存机制**: 系统自动缓存分析结果，减少重复计算
2. **分页查询**: 大数据量时使用分页避免内存溢出
3. **定时清理**: 定期清理历史数据避免数据库过大
4. **并发收集**: 多集群数据并发收集提高效率

## 🔧 故障排除

### 常见问题

#### 1. 无法获取 Metrics 数据
**现象**: 系统显示警告 "无法获取 metrics 数据"
**解决方案**:
```bash
# 检查 metrics-server 是否安装
kubectl get pods -n kube-system | grep metrics-server

# 安装 metrics-server
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# 验证 metrics API
kubectl top nodes
kubectl top pods --all-namespaces
```

#### 2. 集群连接失败
**现象**: 告警显示 "集群连接异常"
**解决方案**:
```bash
# 验证 kubeconfig 文件
kubectl --kubeconfig=/path/to/config get nodes

# 测试网络连通性
curl -k https://your-k8s-api-server:6443/version

# 检查证书有效期
kubectl config view --raw | grep client-certificate-data | awk '{print $2}' | base64 -d | openssl x509 -dates -noout
```

#### 3. 数据库连接问题
**现象**: 应用启动失败，提示数据库连接错误
**解决方案**:
```bash
# 检查配置文件
cat config.toml | grep -A 5 "\[database\]"

# 测试数据库连接
mysql -h localhost -u username -p database_name

# 创建数据库和用户
CREATE DATABASE cluster_resource_insight;
CREATE USER 'app_user'@'%' IDENTIFIED BY 'password';
GRANT ALL PRIVILEGES ON cluster_resource_insight.* TO 'app_user'@'%';
```

#### 4. 前端页面加载问题
**现象**: Web界面显示异常或加载失败
**解决方案**:
```bash
# 重新构建前端
cd web
npm install
npm run build

# 检查静态文件
ls -la web/dist/

# 清除浏览器缓存
# 或使用无痕模式访问
```

#### 5. 内存使用过高
**现象**: 应用内存占用持续增长
**解决方案**:
```bash
# 清理历史数据
curl -X DELETE "http://localhost:9999/api/v1/history/cleanup?retention_days=7"

# 清理活动数据
curl -X DELETE "http://localhost:9999/api/v1/activities/cleanup?retention_days=3"

# 检查数据库大小
SELECT table_name, 
       ROUND(((data_length + index_length) / 1024 / 1024), 2) AS 'Size(MB)'
FROM information_schema.tables 
WHERE table_schema = 'cluster_resource_insight';
```

## 📞 支持与贡献

### 获取帮助
- 🐛 **Bug报告**: 请在 GitHub Issues 中提交
- 💡 **功能建议**: 欢迎提交 Feature Request
- 📖 **文档问题**: 帮助改进文档说明

### 贡献指南
1. Fork 项目到个人仓库
2. 创建功能分支: `git checkout -b feature/amazing-feature`
3. 提交修改: `git commit -m 'Add amazing feature'`
4. 推送分支: `git push origin feature/amazing-feature`  
5. 提交 Pull Request

### 开发环境
```bash
# 开发模式启动后端
go run cmd/main.go

# 开发模式启动前端
cd web
npm run dev

# 代码格式化
go fmt ./...
npm run lint

# 运行测试
go test ./...
npm run test
```

## 📄 许可证

MIT License

Copyright (c) 2025 ClusterResourceInsight

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

---

**ClusterResourceInsight** - 让 Kubernetes 集群资源管理更智能、更高效 🚀