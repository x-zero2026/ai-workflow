# AI Workflow - Go 本地测试服务

用于本地开发和测试的 Go 服务器

## 🎯 用途

- ✅ 快速开发和迭代
- ✅ 本地调试
- ✅ API 测试
- ✅ 集成测试

## 📁 目录结构

```
go/
├── cmd/                 # 命令行工具
│   └── server/         # 服务器入口
│       └── main.go
│
├── pkg/                # 共享包
│   ├── workflow/       # 工作流逻辑
│   ├── auth/          # 认证相关
│   ├── db/            # 数据库操作
│   └── models/        # 数据模型
│
├── api/                # API 处理器
│   ├── handlers/      # HTTP 处理器
│   └── middleware/    # 中间件
│
├── internal/           # 内部包
│   └── config/        # 配置管理
│
├── .env.example       # 环境变量示例
├── .env               # 环境变量（不提交）
├── go.mod             # Go 模块定义
├── go.sum             # 依赖锁定
└── README.md          # 本文件
```

## 🚀 快速开始

### 1. 初始化项目

```bash
# 初始化 Go 模块
go mod init github.com/x-zero2026/ai-workflow

# 安装依赖
go get github.com/lib/pq              # PostgreSQL 驱动
go get github.com/golang-jwt/jwt/v5   # JWT
go get github.com/joho/godotenv       # 环境变量
```

### 2. 配置环境变量

```bash
# 复制示例文件
cp .env.example .env

# 编辑配置
vim .env
```

**.env 内容**:
```env
# 服务器配置
PORT=8080
HOST=localhost

# 数据库配置
SUPABASE_URL=https://xxx.supabase.co
DB_PASSWORD=your-password

# 认证配置
JWT_SECRET=your-jwt-secret
DID_LOGIN_API=https://xxx.execute-api.us-east-1.amazonaws.com/prod

# 日志配置
LOG_LEVEL=debug
```

### 3. 运行服务器

```bash
# 开发模式（自动重载需要安装 air）
air

# 或直接运行
go run cmd/server/main.go

# 访问
curl http://localhost:8080/health
```

## 📝 创建基本结构

### 主程序 (cmd/server/main.go)

```go
package main

import (
    "log"
    "net/http"
    "os"

    "github.com/joho/godotenv"
)

func main() {
    // 加载环境变量
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    // 设置路由
    http.HandleFunc("/health", healthHandler)
    http.HandleFunc("/api/workflows", workflowsHandler)

    log.Printf("Server starting on port %s", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatal(err)
    }
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"status":"ok"}`))
}

func workflowsHandler(w http.ResponseWriter, r *http.Request) {
    // TODO: 实现工作流处理逻辑
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"workflows":[]}`))
}
```

### 数据库连接 (pkg/db/postgres.go)

```go
package db

import (
    "database/sql"
    "fmt"
    "os"

    _ "github.com/lib/pq"
)

func Connect() (*sql.DB, error) {
    supabaseURL := os.Getenv("SUPABASE_URL")
    dbPassword := os.Getenv("DB_PASSWORD")

    // 构建连接字符串
    connStr := fmt.Sprintf(
        "postgresql://postgres.xxx:%s@aws-1-ap-south-1.pooler.supabase.com:6543/postgres",
        dbPassword,
    )

    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }

    if err := db.Ping(); err != nil {
        return nil, err
    }

    return db, nil
}
```

### 认证中间件 (api/middleware/auth.go)

```go
package middleware

import (
    "net/http"
    "strings"

    "github.com/x-zero2026/ai-workflow/pkg/auth"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 获取 Token
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing authorization header", http.StatusUnauthorized)
            return
        }

        // 验证 Token
        token := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := auth.ValidateToken(token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // 将用户信息添加到上下文
        // TODO: 使用 context 传递用户信息

        next(w, r)
    }
}
```

## 🔧 开发工具

### Air - 热重载

```bash
# 安装 air
go install github.com/cosmtrek/air@latest

# 创建配置文件
cat > .air.toml << 'EOF'
root = "."
tmp_dir = "tmp"

[build]
cmd = "go build -o ./tmp/main ./cmd/server"
bin = "tmp/main"
include_ext = ["go"]
exclude_dir = ["tmp", "vendor"]
EOF

# 运行
air
```

### 调试

使用 VS Code 调试配置 (`.vscode/launch.json`):

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Server",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "${workspaceFolder}/cmd/server",
            "env": {
                "PORT": "8080"
            }
        }
    ]
}
```

## 🧪 测试

### 单元测试

```bash
# 运行所有测试
go test ./...

# 运行特定包
go test ./pkg/workflow

# 查看覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### API 测试

```bash
# 健康检查
curl http://localhost:8080/health

# 创建工作流（需要 Token）
curl -X POST http://localhost:8080/api/workflows \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Workflow",
    "description": "Test workflow"
  }'
```

## 📊 性能分析

```bash
# CPU 分析
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# 内存分析
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

## 🔍 日志

```go
import "log"

// 基本日志
log.Println("Info message")
log.Printf("User %s created workflow", username)

// 错误日志
log.Printf("Error: %v", err)

// 致命错误（会退出程序）
log.Fatal("Fatal error")
```

## 📦 构建

```bash
# 开发构建
go build -o bin/server cmd/server/main.go

# 生产构建（优化）
go build -ldflags="-s -w" -o bin/server cmd/server/main.go

# 交叉编译（Linux）
GOOS=linux GOARCH=amd64 go build -o bin/server-linux cmd/server/main.go
```

## 🔐 环境变量

创建 `.env.example`:

```env
# 服务器配置
PORT=8080
HOST=localhost

# 数据库配置
SUPABASE_URL=https://your-project.supabase.co
DB_PASSWORD=your-password

# 认证配置
JWT_SECRET=your-jwt-secret
DID_LOGIN_API=https://your-api.execute-api.us-east-1.amazonaws.com/prod

# 日志配置
LOG_LEVEL=debug
```

## 🆘 常见问题

### Q: 如何连接数据库？

**A**: 使用 Supabase Pooler 连接（端口 6543），参考 `pkg/db/postgres.go`

### Q: 如何验证 JWT Token？

**A**: 使用与 DID Login 相同的 JWT_SECRET，参考 `pkg/auth/jwt.go`

### Q: 如何处理 CORS？

**A**: 添加 CORS 中间件：

```go
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next(w, r)
    }
}
```

## 📚 相关文档

- [后端总览](../README.md)
- [Lambda 部署](../lambda/README.md)
- [DID Login 集成](../../../did-login-lambda/README.md)

## 🎯 下一步

1. 创建基本项目结构
2. 实现数据库连接
3. 实现 JWT 认证
4. 创建工作流 API
5. 编写测试
6. 准备迁移到 Lambda
