# AI Workflow Center - Frontend

React 前端应用，用于管理和执行 AI 工作流（支持 Coze 和 n8n）。

## 🚀 快速开始

### 本地开发

```bash
# 安装依赖
npm install

# 配置环境变量
cp .env.example .env
# 编辑 .env 设置 API URLs

# 启动开发服务器
npm run dev
```

访问 http://localhost:5174

### 构建生产版本

```bash
npm run build
```

## 📦 部署到 AWS Amplify

本项目已配置 `amplify.yml` 和 `.amplifyignore`，支持自动部署并优化流量使用。

### 环境变量

在 Amplify 控制台配置：

```
VITE_API_BASE_URL = https://ynnid7kam5.execute-api.us-east-1.amazonaws.com/prod
VITE_LOGIN_API_BASE_URL = https://ynnid7kam5.execute-api.us-east-1.amazonaws.com/prod
VITE_DID_LOGIN_URL = https://main.d2fozf421c6ftf.amplifyapp.com/dashboard
```

### 部署优化

- ✅ 使用 `.amplifyignore` 排除文档和测试文件
- ✅ 使用 `npm ci` 加速依赖安装
- ✅ 缓存 `node_modules` 加速构建
- ✅ 只部署 `dist/` 目录的构建产物

### 详细文档

完整的部署指南和文档请查看：
- `../` - AI Workflow 项目文档
- `../../docs/` - 通用文档

## 📚 技术栈

- React 18
- React Router 6
- Vite
- Axios

## 🎨 特性

- ✅ 工作流管理（创建、编辑、删除）
- ✅ 工作流执行（Coze、n8n）
- ✅ 模板工作流
- ✅ 工作流分享
- ✅ 项目隔离
- ✅ 流畅的动画效果

## 📄 许可证

MIT License
