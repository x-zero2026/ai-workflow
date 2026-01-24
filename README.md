# AI Workflow Platform

AI 工作流平台 - 基于 DID Login 的 AI 工作流管理系统

## 📁 项目结构

```
ai-workflow/
├── backend/              # 后端服务
│   ├── go/              # Go 本地测试服务
│   └── lambda/          # AWS Lambda 部署
│
└── frontend/            # 前端应用
```

## 🚀 快速开始

### 后端开发

#### 本地测试
```bash
cd backend/go
go run main.go
# 访问 http://localhost:8080
```

#### 部署到 Lambda
```bash
cd backend/lambda
sam build
sam deploy --guided
```

### 前端开发

```bash
cd frontend
npm install
npm run dev
# 访问 http://localhost:5173
```

## 🔗 依赖项目

本项目依赖 DID Login 平台进行用户认证：
- [did-login-lambda](../did-login-lambda/) - 认证后端
- [did-login-ui](../did-login-ui/) - 认证前端

## 📚 文档

- [后端 Go 开发](./backend/go/README.md)
- [后端 Lambda 部署](./backend/lambda/README.md)
- [前端开发](./frontend/README.md)

## 🛠️ 技术栈

### 后端
- **语言**: Go 1.21+
- **部署**: AWS Lambda + API Gateway
- **数据库**: PostgreSQL (Supabase)
- **认证**: JWT (来自 DID Login)

### 前端
- **框架**: Vue 3 + Vite
- **路由**: Vue Router
- **HTTP**: Axios
- **部署**: AWS Amplify

## 🎯 功能规划

- [ ] AI 工作流创建和管理
- [ ] 工作流节点编辑器
- [ ] 工作流执行引擎
- [ ] 执行历史和日志
- [ ] 团队协作
- [ ] 模板市场

## 📝 开发状态

🚧 项目初始化中...

## 🔐 环境变量

### 后端
```bash
# backend/go/.env
SUPABASE_URL=https://xxx.supabase.co
DB_PASSWORD=your-password
JWT_SECRET=your-secret
DID_LOGIN_API=https://xxx.execute-api.us-east-1.amazonaws.com/prod
```

### 前端
```bash
# frontend/.env
VITE_API_BASE_URL=http://localhost:8080
VITE_DID_LOGIN_URL=https://main.xxx.amplifyapp.com
```

## 🆘 需要帮助？

查看各子目录的 README 文件获取详细信息。
