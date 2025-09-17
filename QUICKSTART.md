

# K8s 多集群资源监控系统 - 快速开始指南

## 系统概述

这是一个企业级的 K8s 多集群资源监控分析平台，支持：
- 多个 K8s 集群的统一管理
- 集群连接配置的安全存储（数据库加密）
- Pod 资源配置分析和优化建议
- 定时监控和历史数据存储
- Web 界面管理和 RESTful API

## 快速开始

### 1. 环境准备

确保您已安装：
- Go 1.24+
- MySQL 5.7+ 或 8.0+
- kubectl（如果要连接 K8s 集群）

### 2. 数据库准备

```sql
-- 创建数据库
CREATE DATABASE cluster_resource_insight CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 可选：创建专用用户
CREATE USER 'cluster_user'@'localhost' IDENTIFIED BY 'your_password';
GRANT ALL PRIVILEGES ON cluster_resource_insight.* TO 'cluster_user'@'localhost';
FLUSH PRIVILEGES;
```

### 3. 配置应用

编辑 `config.toml` 文件：

```toml
[database]
host = "localhost"
port = 3306
username = "root"
password = "your_mysql_password"  # 或通过环境变量 DB_PASSWORD 设置
database = "cluster_resource_insight"
```

### 4. 初始化系统

```bash
# 首次运行，执行数据库迁移
./start.sh migrate

# 启动应用
./start.sh run
```

### 5. 访问系统

- **资源监控**: http://localhost:8080
- **集群管理**: http://localhost:8080/clusters  
- **API 文档**: http://localhost:8080/api/v1/health

## 集群管理

### 添加 K8s 集群

#### 方法1：通过 API

```bash
# 使用 Token 认证
curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "cluster_name": "prod-cluster",
    "cluster_alias": "生产环境",
    "api_server": "https://k8s-api.example.com:6443",
    "auth_type": "token",
    "auth_config": {
      "bearer_token": "your_k8s_token_here"
    },
    "tags": ["production", "asia"],
    "collect_interval": 30
  }'

# 使用 Kubeconfig 认证
curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "cluster_name": "dev-cluster",
    "cluster_alias": "开发环境",
    "api_server": "https://dev-k8s.example.com:6443",
    "auth_type": "kubeconfig",
    "auth_config": {
      "kubeconfig": "apiVersion: v1\nkind: Config\n..."
    },
    "collect_interval": 60
  }'
```

#### 方法2：通过 Web 界面

访问 http://localhost:8080/clusters 使用图形界面管理集群。

### 测试集群连接

```bash
# 测试指定集群
curl -X POST http://localhost:8080/api/v1/clusters/1/test

# 批量测试所有集群
curl -X POST http://localhost:8080/api/v1/clusters/batch-test

# 创建前测试配置
curl -X POST http://localhost:8080/api/v1/clusters/test \
  -H "Content-Type: application/json" \
  -d '{集群配置JSON}'
```

## 高级配置

### 环境变量

```bash
# 数据库密码（推荐用于生产环境）
export DB_PASSWORD=your_password

# 自定义加密密钥（32字节）
export ENCRYPTION_KEY=your_32_byte_encryption_key_here

# 使用自定义配置文件
CONFIG_FILE=/path/to/custom.toml ./start.sh run
```

### 配置文件说明

```toml
# 应用配置
[app]
port = 8080      # Web 服务端口
debug = false    # 调试模式

# 监控配置
[monitoring]
default_collect_interval = 30     # 默认采集间隔（分钟）
max_concurrent_collections = 10   # 最大并发采集数
data_retention_days = 30          # 历史数据保留天数

# 告警阈值
[alert]
memory_usage_threshold_low = 20   # 内存利用率过低阈值
cpu_usage_threshold_low = 15      # CPU利用率过低阈值
```

## API 接口

### 集群管理 API

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/clusters` | 获取所有集群 |
| POST | `/api/v1/clusters` | 创建集群 |
| GET | `/api/v1/clusters/{id}` | 获取指定集群 |
| PUT | `/api/v1/clusters/{id}` | 更新集群配置 |
| DELETE | `/api/v1/clusters/{id}` | 删除集群 |
| POST | `/api/v1/clusters/{id}/test` | 测试集群连接 |
| POST | `/api/v1/clusters/test` | 测试集群配置 |
| POST | `/api/v1/clusters/batch-test` | 批量测试集群 |

### 资源分析 API

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/analysis` | 获取资源分析结果 |
| GET | `/api/v1/pods` | 获取 Pod 数据 |
| GET | `/api/v1/health` | 健康检查 |

## 故障排除

### 常见问题

1. **数据库连接失败**
   ```
   检查配置文件中的数据库连接信息
   确保 MySQL 服务正在运行
   验证数据库用户权限
   ```

2. **集群连接失败**
   ```
   检查 API Server 地址是否正确
   验证认证信息（Token/证书/Kubeconfig）
   确认网络连通性
   检查防火墙设置
   ```

3. **Metrics API 不可用**
   ```
   这是正常情况，如果集群没有安装 metrics-server
   系统会跳过实际使用量监控，只分析配置问题
   ```

### 日志查看

```bash
# 查看应用日志
tail -f logs/app.log

# 查看实时日志（如果配置了）
./bin/cluster-resource-insight -config=config.toml
```

### 数据备份

```bash
# 备份数据库
mysqldump -u root -p cluster_resource_insight > backup.sql

# 恢复数据库
mysql -u root -p cluster_resource_insight < backup.sql
```

## 安全考虑

1. **加密存储**: 所有集群认证信息都经过 AES-256 加密存储
2. **密钥管理**: 生产环境请更换默认加密密钥
3. **网络安全**: 建议使用 HTTPS 和防火墙保护
4. **权限控制**: 数据库用户使用最小权限原则

## 下一步

- 配置定时采集任务
- 设置告警规则
- 创建自定义报告
- 集成监控系统

## 技术支持

如有问题，请检查：
1. 日志文件：`logs/app.log`
2. 配置文件：`config.toml`
3. 数据库连接状态
4. K8s 集群网络连通性