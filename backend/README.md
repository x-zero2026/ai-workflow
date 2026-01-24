# AI Workflow Backend

AI 工作流平台后端服务

## 📁 目录结构

```
backend/
├── go/              # Go 本地测试服务
│   ├── cmd/        # 主程序入口
│   ├── pkg/        # 共享包
│   ├── api/        # API 处理器
│   └── main.go     # 本地服务器
│
└── lambda/          # AWS Lambda 部署
    ├── template.yaml    # SAM 模板
    ├── functions/       # Lambda 函数
    └── shared/          # 共享代码
```

## 🎯 开发流程

### 1. 本地开发和测试

在 `go/` 目录中开发和测试：

```bash
cd go
go run main.go
```

优势：
- ✅ 快速迭代
- ✅ 实时调试
- ✅ 完整的 Go 工具链支持

### 2. 部署到 Lambda

测试通过后，部署到 `lambda/`：

```bash
cd lambda
sam build
sam deploy
```

## 🔄 代码同步

`go/` 和 `lambda/` 共享核心业务逻辑：

- `go/pkg/` → `lambda/shared/` - 共享包
- `go/api/` → `lambda/functions/` - API 处理器

## 📚 详细文档

- [Go 本地开发](./go/README.md)
- [Lambda 部署](./lambda/README.md)

## 🛠️ 技术栈

- **语言**: Go 1.21+
- **框架**: 标准库 + 自定义路由
- **数据库**: PostgreSQL (Supabase)
- **认证**: JWT (DID Login)
- **部署**: AWS Lambda + API Gateway

## 🔐 认证集成

本服务依赖 DID Login 平台进行用户认证：

```go
// 验证 JWT Token
token := r.Header.Get("Authorization")
claims, err := auth.ValidateToken(token)
if err != nil {
    // 未授权
}

// 获取用户信息
userDID := claims.DID
username := claims.Username
```

## 📊 API 端点规划

| 端点 | 方法 | 说明 |
|------|------|------|
| /api/workflows | GET | 列出工作流 |
| /api/workflows | POST | 创建工作流 |
| /api/workflows/{id} | GET | 获取工作流详情 |
| /api/workflows/{id} | PUT | 更新工作流 |
| /api/workflows/{id} | DELETE | 删除工作流 |
| /api/workflows/{id}/execute | POST | 执行工作流 |
| /api/executions | GET | 列出执行历史 |
| /api/executions/{id} | GET | 获取执行详情 |

## 🚀 快速开始

### 初始化 Go 项目

```bash
cd go
go mod init github.com/x-zero2026/ai-workflow
go mod tidy
```

### 创建基本结构

```bash
mkdir -p cmd/server pkg/workflow pkg/auth api/handlers
touch main.go
```

### 运行本地服务器

```bash
go run main.go
```

## 📝 开发规范

### 代码组织

- `cmd/` - 可执行程序入口
- `pkg/` - 可复用的包
- `api/` - HTTP 处理器
- `internal/` - 内部包（不对外暴露）

### 错误处理

```go
if err != nil {
    log.Printf("Error: %v", err)
    http.Error(w, "Internal server error", http.StatusInternalServerError)
    return
}
```

### 响应格式

```go
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}
```

## 🔍 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./pkg/workflow

# 查看覆盖率
go test -cover ./...
```

## 📦 依赖管理

```bash
# 添加依赖
go get github.com/lib/pq

# 更新依赖
go get -u ./...

# 清理未使用的依赖
go mod tidy
```

## 🆘 需要帮助？

- 查看 [go/README.md](./go/README.md) - 本地开发指南
- 查看 [lambda/README.md](./lambda/README.md) - Lambda 部署指南
