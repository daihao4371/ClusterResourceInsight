# 用户管理、角色管理、API管理、审计日志功能接口文档

## 1. 用户管理接口 `/api/users`

### 1.1 创建用户
- **接口**: `POST /api/users`
- **请求体**:
```json
{
  "username": "testuser",
  "password": "password123",
  "email": "test@example.com",
  "real_name": "测试用户",
  "phone": "13800138000",
  "role_ids": [1, 2]
}
```

### 1.2 获取用户列表
- **接口**: `GET /api/users?page=1&page_size=10&keyword=test&status=active`
- **参数**:
  - `page`: 页码（默认1）
  - `page_size`: 每页数量（默认10，最大100）
  - `keyword`: 搜索关键字（支持用户名、真实姓名、邮箱）
  - `status`: 状态筛选（active/inactive/locked）

### 1.3 获取用户详情
- **接口**: `GET /api/users/:id`

### 1.4 更新用户信息
- **接口**: `PUT /api/users/:id`
- **请求体**:
```json
{
  "email": "newemail@example.com",
  "real_name": "新名称",
  "phone": "13900139000",
  "status": "active",
  "role_ids": [1]
}
```

### 1.5 删除用户
- **接口**: `DELETE /api/users/:id`

### 1.6 修改密码
- **接口**: `PUT /api/users/:id/password`
- **请求体**:
```json
{
  "old_password": "oldpass123",
  "new_password": "newpass123"
}
```

### 1.7 重置密码（管理员）
- **接口**: `PUT /api/users/:id/reset-password`
- **请求体**:
```json
{
  "new_password": "resetpass123"
}
```

---

## 2. 角色管理接口 `/api/roles`

### 2.1 创建角色
- **接口**: `POST /api/roles`
- **请求体**:
```json
{
  "role_name": "开发者",
  "role_code": "developer",
  "description": "开发者角色",
  "permissions": ["read", "write", "deploy"]
}
```

### 2.2 获取角色列表
- **接口**: `GET /api/roles?page=1&page_size=10&keyword=dev&status=active`
- **参数**:
  - `page`: 页码
  - `page_size`: 每页数量
  - `keyword`: 搜索关键字（支持角色名称、编码、描述）
  - `status`: 状态筛选

### 2.3 获取所有激活角色（用于选择器）
- **接口**: `GET /api/roles/active`

### 2.4 获取角色详情
- **接口**: `GET /api/roles/:id`

### 2.5 更新角色
- **接口**: `PUT /api/roles/:id`
- **请求体**:
```json
{
  "role_name": "高级开发者",
  "description": "高级开发者角色",
  "permissions": ["*"],
  "status": "active"
}
```

### 2.6 删除角色
- **接口**: `DELETE /api/roles/:id`

### 2.7 获取角色下的用户列表
- **接口**: `GET /api/roles/:id/users?page=1&page_size=10`

### 2.8 获取角色权限
- **接口**: `GET /api/roles/:id/permissions`

---

## 3. API管理接口 `/api/apis`

### 3.1 创建API
- **接口**: `POST /api/apis`
- **请求体**:
```json
{
  "path": "/api/users",
  "method": "GET",
  "group": "用户管理",
  "description": "获取用户列表",
  "permission": "user:read",
  "is_public": false
}
```

### 3.2 获取API列表
- **接口**: `GET /api/apis?page=1&page_size=10&keyword=user&group=用户管理&method=GET&status=active&is_public=false`
- **参数**:
  - `page`: 页码
  - `page_size`: 每页数量
  - `keyword`: 搜索关键字（路径、描述）
  - `group`: 分组筛选
  - `method`: HTTP方法筛选
  - `status`: 状态筛选
  - `is_public`: 是否公开接口

### 3.3 获取API分组列表
- **接口**: `GET /api/apis/groups`

### 3.4 获取API统计信息
- **接口**: `GET /api/apis/statistics`

### 3.5 获取API详情
- **接口**: `GET /api/apis/:id`

### 3.6 更新API
- **接口**: `PUT /api/apis/:id`
- **请求体**:
```json
{
  "description": "更新后的描述",
  "permission": "user:update",
  "status": "active"
}
```

### 3.7 删除API
- **接口**: `DELETE /api/apis/:id`

### 3.8 批量创建API
- **接口**: `POST /api/apis/batch`
- **请求体**:
```json
{
  "apis": [
    {
      "path": "/api/users",
      "method": "GET",
      "group": "用户管理",
      "description": "获取用户列表",
      "permission": "user:read",
      "is_public": false
    }
  ]
}
```

### 3.9 同步API
- **接口**: `POST /api/apis/sync`
- **说明**: 同步系统中的API，更新已有的，创建缺失的

---

## 4. 审计日志接口 `/api/audit`

### 4.1 查询审计日志
- **接口**: `GET /api/audit/logs?page=1&page_size=10&user_id=1&username=admin&operate_type=create&module=用户管理&resource=user&start_time=2025-01-01 00:00:00&end_time=2025-12-31 23:59:59&min_duration=0&max_duration=1000&response_code=200`
- **参数**:
  - `page`: 页码
  - `page_size`: 每页数量
  - `user_id`: 用户ID筛选
  - `username`: 用户名筛选
  - `operate_type`: 操作类型（create/update/delete/query/login/logout）
  - `module`: 模块筛选
  - `resource`: 资源筛选
  - `start_time`: 开始时间（格式：2006-01-02 15:04:05）
  - `end_time`: 结束时间
  - `min_duration`: 最小耗时（毫秒）
  - `max_duration`: 最大耗时（毫秒）
  - `response_code`: 响应码筛选

### 4.2 获取审计日志详情
- **接口**: `GET /api/audit/logs/:id`

### 4.3 获取指定用户的审计日志
- **接口**: `GET /api/audit/users/:user_id/logs?page=1&page_size=10`

### 4.4 获取审计统计信息
- **接口**: `GET /api/audit/statistics?start_time=2025-01-01 00:00:00&end_time=2025-12-31 23:59:59`
- **响应示例**:
```json
{
  "total": 1000,
  "by_operate_type": [
    {"operate_type": "query", "count": 600},
    {"operate_type": "create", "count": 200}
  ],
  "by_module": [
    {"module": "用户管理", "count": 500}
  ],
  "by_user": [
    {"username": "admin", "count": 300}
  ],
  "by_status": [
    {"response_code": 200, "count": 900}
  ],
  "avg_duration": 150.5,
  "slow_request_count": 10,
  "error_count": 20
}
```

### 4.5 获取操作趋势
- **接口**: `GET /api/audit/trend?days=7`
- **参数**:
  - `days`: 统计天数（默认7天，最大90天）

### 4.6 清理旧审计日志
- **接口**: `DELETE /api/audit/cleanup?days=90`
- **参数**:
  - `days`: 保留天数（最少30天）

### 4.7 导出审计日志
- **接口**: `GET /api/audit/export?user_id=1&operate_type=create&module=用户管理&start_time=2025-01-01 00:00:00&end_time=2025-12-31 23:59:59`
- **说明**: 导出CSV格式的审计日志

---

## 5. 默认数据

### 5.1 默认角色
系统初始化时会自动创建以下角色:
- **超级管理员** (admin): 拥有所有权限 `["*"]`
- **普通用户** (user): 只有查看权限 `["read"]`

### 5.2 默认用户
- **用户名**: admin
- **密码**: admin123
- **邮箱**: admin@example.com
- **角色**: 超级管理员

⚠️ **重要**: 首次登录后请立即修改默认密码！

---

## 6. 数据模型说明

### 6.1 用户状态
- `active`: 正常激活
- `inactive`: 已禁用
- `locked`: 已锁定

### 6.2 操作类型
- `create`: 创建
- `update`: 更新
- `delete`: 删除
- `query`: 查询
- `login`: 登录
- `logout`: 登出

### 6.3 HTTP方法
- `GET`: 查询操作
- `POST`: 创建操作
- `PUT`: 更新操作
- `DELETE`: 删除操作
- `PATCH`: 部分更新

---

## 7. 使用示例

### 7.1 创建用户完整流程
```bash
# 1. 创建角色
curl -X POST http://localhost:8080/api/roles \
  -H "Content-Type: application/json" \
  -d '{
    "role_name": "开发者",
    "role_code": "developer",
    "description": "开发者角色",
    "permissions": ["read", "write"]
  }'

# 2. 创建用户并关联角色
curl -X POST http://localhost:8080/api/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "developer1",
    "password": "dev123456",
    "email": "dev1@example.com",
    "real_name": "开发者1号",
    "role_ids": [2]
  }'

# 3. 查询用户列表
curl http://localhost:8080/api/users?page=1&page_size=10

# 4. 查询审计日志
curl http://localhost:8080/api/audit/logs?operate_type=create&module=用户管理
```

---

## 8. 注意事项

1. **密码安全**: 所有密码使用 bcrypt 加密存储，不可逆
2. **软删除**: 用户、角色、API 都采用软删除机制
3. **审计日志**: 所有操作自动记录审计日志，敏感信息自动脱敏
4. **权限控制**: 系统内置角色不允许删除和修改核心属性
5. **角色使用检查**: 删除角色前会检查是否有用户在使用
6. **分页限制**: 每页最多返回100条记录
7. **时间格式**: 统一使用 `2006-01-02 15:04:05` 格式