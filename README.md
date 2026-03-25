# GoChat - 即时通讯系统

![License](https://img.shields.io/badge/License-MIT-green.svg)
![Backend](https://img.shields.io/badge/Backend-Go%20%7C%20Gin-blue)
![Frontend](https://img.shields.io/badge/Frontend-Vue3%20%7C%20TypeScript-42b883)

GoChat 是一个前后端分离的即时通讯系统示例项目，支持账号认证、单聊、群聊、联系人管理、文件上传、在线状态同步与实时消息推送。

系统采用 `Go + Gin + GORM + MySQL + WebSocket` 构建后端服务，前端基于 `Vue 3 + Pinia + Element Plus + Vite` 实现，适合作为 IM 类项目的学习与二次开发基础。

## ✨ 项目特色

- 实时通信：基于 WebSocket 的消息收发、在线状态广播与消息确认。
- 聊天能力完整：支持单聊、群聊、会话列表、历史消息拉取。
- 用户与社交关系：支持注册登录、资料维护、好友搜索、好友申请、好友管理。
- 资源上传：支持头像、聊天图片、聊天文件、语音文件上传。
- 前后端分离：后端 API 与前端 SPA 解耦，便于独立开发与部署。
- 易扩展：模块化目录结构，服务、处理器、WebSocket 路由职责清晰。

## 📸 系统预览

你可以在前端启动后访问聊天页面，主要包含以下模块：

- 登录/注册页面
- 会话列表与消息面板
- 通讯录与好友管理
- 群聊管理与成员管理
- 文件与语音消息交互

## 🔧 技术栈

### 后端

- 语言：Go
- 框架：Gin
- ORM：GORM
- 数据库：MySQL
- 实时通信：WebSocket (gorilla/websocket)
- 认证：JWT

### 前端

- 框架：Vue 3 + TypeScript
- 构建工具：Vite
- 状态管理：Pinia
- UI 组件库：Element Plus

## 🚀 快速开始

请按以下步骤在本地运行项目。

### 1. 克隆项目

```bash
git clone <https://github.com/xg-debug/gochat.git>
cd gochat
```

### 2. 准备数据库

项目提供了建表 SQL 文件：

- `backend/sql/schema.sql`

在 MySQL 中创建数据库后，导入该文件：

```sql
source backend/sql/schema.sql;
```

### 3. 配置后端

编辑配置文件：

- `backend/configs/config.yaml`

请根据你的本地环境配置 MySQL、服务端口、JWT 等参数。

### 4. 启动后端服务

```bash
cd backend
go mod tidy
go run ./cmd
```

默认会按 `config.yaml` 中的配置监听服务地址。

### 5. 启动前端服务

```bash
cd frontend
npm install
npm run dev
```

启动后在浏览器打开终端提示地址（通常为 `http://localhost:5173`）。

## 📂 项目结构

### 高层结构

```text
gochat/
├── backend/                  # Go 后端服务
│   ├── cmd/                  # 程序入口
│   ├── configs/              # 配置文件
│   ├── internal/             # 核心业务代码
│   └── sql/                  # 数据库 SQL
├── frontend/                 # Vue 前端应用
└── README.md
```

### 后端核心结构

```text
backend/internal/
├── config/                   # 配置加载
├── dto/                      # 请求与响应 DTO
├── handler/                  # HTTP/WS 处理层
├── model/                    # 数据模型
├── pkg/                      # 通用能力（auth/db/logger）
├── service/                  # 业务服务层
└── ws/                       # WebSocket Hub 与路由
```

### 前端核心结构

```text
frontend/src/
├── components/               # UI 组件（会话、消息、联系人等）
├── router/                   # 路由
├── services/                 # API 与 WebSocket 客户端
├── stores/                   # Pinia 状态管理
├── types/                    # 类型定义
├── utils/                    # 工具函数
└── views/                    # 页面视图
```

## 🚀 后端核心系统详解

### 核心架构

系统分为以下几个关键层：

- API 层（`handler`）：
  - 提供登录注册、用户资料、好友、群聊、上传、消息等 HTTP 接口。
  - 提供 WebSocket 接入用于实时消息收发。
- 业务层（`service`）：
  - 封装用户、好友、群组、上传等业务逻辑。
- 数据层（`model + gorm`）：
  - 使用 GORM 映射 MySQL 表结构。
- 实时通信层（`ws`）：
  - 负责连接管理、消息分发、在线状态广播与信令消息路由。

### 关键模块说明

- `backend/cmd/main.go`
  - 后端启动入口，初始化配置、数据库、路由与 WebSocket Hub。

- `backend/internal/handler/`
  - 接口处理层，对外暴露 REST API 与 WS 接入点。

- `backend/internal/service/`
  - 业务核心层，处理账号、会话、好友、群组与上传逻辑。

- `backend/internal/ws/`
  - WebSocket 核心，管理在线连接、消息路由、回执与在线状态。

- `backend/internal/pkg/auth/`
  - JWT 鉴权、中间件与 token 处理。

## 🧩 前端模块说明

- `views/ChatView.vue`
  - 聊天主页面，整合会话、消息、联系人、群组与通话面板。

- `stores/chat.ts`
  - 聊天状态管理，负责会话加载、消息缓存、WS 连接、消息发送/接收。

- `services/api.ts`
  - HTTP 请求封装，统一调用后端 API。

- `services/ws.ts`
  - WebSocket 客户端封装，处理重连、心跳、消息分发。

## 🔐 API 概览

后端主要接口前缀为 `/api`，包括：

- 认证：`/login`、`/register`、`/logout`
- 用户：`/profile`、`/upload/avatar`
- 好友：`/user/search`、`/friend/request`、`/friend/handle`、`/contacts`
- 会话：`/conversations`、`/conversations/search`、`/messages`
- 群聊：`/group/create`、`/group/join`、`/group/profile`、`/group/members` 等
- 上传：`/upload/chat/image`、`/upload/chat/file`、`/upload/chat/audio`、`/upload/group/avatar`
- WebSocket：`/ws`（需携带 token）

## 📜 部署

你可以根据自身环境将前后端分开部署：

- 后端：以 systemd / Docker / 进程管理器运行 Go 服务。
- 前端：`npm run build` 后将 `dist` 部署到 Nginx 等静态服务器。
- 数据库：MySQL 独立部署，按需配置备份策略。

## 🙏 致谢

感谢开源社区提供的 Go、Vue、Gin、GORM、Element Plus 等优秀项目。

## 📄 许可证

本项目采用 MIT 许可证，详见 `LICENSE`。
