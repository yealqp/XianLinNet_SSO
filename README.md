# OAuth Server

[English](#english) | [中文](#中文)

---

## English

A full-featured OAuth 2.0 / OpenID Connect server built with Go (Beego) and Vue 3, supporting user authentication, authorization, role-based access control (RBAC), and real-name verification.

### Features

- 🔐 OAuth 2.0 & OpenID Connect support
- 👥 User management with role-based access control (RBAC)
- 🎫 Token management and validation
- 📧 Email verification and password reset
- 🔑 JWT-based authentication
- 🎨 Modern Vue 3 + Ant Design Vue frontend
- 💾 PostgreSQL database support
- 🚀 Redis caching support (optional)
- 🔒 RSA encryption for sensitive data
- 🌐 CORS support for cross-origin requests
- 📱 Real-name verification API integration (optional)

### Tech Stack

**Backend:**
- Go 1.23+
- Beego v2 web framework
- XORM for database operations
- JWT for token management
- Redis for caching (optional)

**Frontend:**
- Vue 3 with TypeScript
- Ant Design Vue 4
- Vite build tool
- Pinia for state management
- Vue Router for routing

### Prerequisites

- Go 1.23 or higher
- Node.js 18+ and pnpm
- PostgreSQL 12 or higher
- Redis (optional, for caching)

### Quick Start

#### 1. Clone the repository

```bash
git clone <repository-url>
cd oauth-server
```

#### 2. Backend Setup

```bash
# Install Go dependencies
go mod download

# Setup PostgreSQL database
# See docs/postgresql-setup.md for detailed instructions
createdb oauth_server

# Copy configuration file
copy .env.example .env

# Edit .env and configure:
# - PostgreSQL connection string (REQUIRED)
# - JWT secret (minimum 32 characters)
# - Admin credentials (REQUIRED for first run)
# - SMTP settings (if email verification needed)
# - Redis settings (optional)

# Initialize database and create admin user
go run main.go init

# Start the backend server
go run main.go
```

The backend server will start on `http://localhost:8080` (or your configured port).

#### 3. Frontend Setup

```bash
cd frontend

# Install dependencies
pnpm install

# Start development server
pnpm dev
```

The frontend will start on `http://localhost:5173`.

For production build:

```bash
pnpm build
```

### Configuration

Edit `.env` to configure:

**Required Settings:**
- Database configuration (PostgreSQL):
  - `dbHost` - PostgreSQL server address (default: localhost)
  - `dbPort` - PostgreSQL port (default: 5432)
  - `dbUser` - Database username
  - `dbPassword` - Database password
  - `dbName` - Database name
  - `dbSSLMode` - SSL mode (disable/require/verify-ca/verify-full)
- `adminEmail`, `adminPassword`, `adminUsername` - Initial admin user credentials
- `jwtSecret` - JWT signing key (minimum 32 characters)

**Database Configuration Example:**
```ini
dbHost = localhost
dbPort = 5432
dbUser = postgres
dbPassword = password
dbName = oauth_server
dbSSLMode = disable
```

See `docs/postgresql-setup.md` for detailed setup instructions

**Optional Settings:**
- Redis cache configuration
- SMTP email settings
- CORS and origin settings
- Real-name verification API

### Project Structure

```
.
├── controllers/       # HTTP request handlers
├── models/           # Database models and operations
├── services/         # Business logic layer
├── routers/          # Route definitions
├── keys/             # RSA key pairs for encryption
├── frontend/         # Vue 3 frontend application
│   ├── src/
│   │   ├── api/      # API client
│   │   ├── components/  # Vue components
│   │   ├── views/    # Page views
│   │   ├── router/   # Route configuration
│   │   └── stores/   # Pinia stores
│   └── dist/         # Production build output
└── main.go           # Application entry point
```

### API Endpoints

- `POST /api/auth/login` - User login
- `POST /api/auth/register` - User registration
- `POST /api/auth/logout` - User logout
- `GET /api/auth/authorize` - OAuth authorization endpoint
- `POST /api/auth/token` - OAuth token endpoint
- `GET /api/admin/*` - Admin management endpoints
- `GET /api/user/*` - User profile endpoints

### Default Admin Account

After running `go run main.go init`, an admin account will be created using the credentials from `.env`:

- Email: As configured in `adminEmail`
- Password: As configured in `adminPassword`
- Username: As configured in `adminUsername`

**Important:** Change these default credentials immediately after first login!

### Development

```bash
# Run backend with hot reload (requires air or similar tool)
go run main.go

# Run frontend development server
cd frontend && pnpm dev

# Build frontend for production
cd frontend && pnpm build
```

### License

Apache License 2.0

---

## 中文

一个功能完整的 OAuth 2.0 / OpenID Connect 服务器，使用 Go (Beego) 和 Vue 3 构建，支持用户认证、授权、基于角色的访问控制 (RBAC) 和实名验证。

### 功能特性

- 🔐 支持 OAuth 2.0 和 OpenID Connect
- 👥 用户管理与基于角色的访问控制 (RBAC)
- 🎫 令牌管理和验证
- 📧 邮箱验证和密码重置
- 🔑 基于 JWT 的身份认证
- 🎨 现代化的 Vue 3 + Ant Design Vue 前端
- 💾 PostgreSQL 数据库支持
- 🚀 Redis 缓存支持（可选）
- 🔒 敏感数据 RSA 加密
- 🌐 支持跨域请求 (CORS)
- 📱 实名验证 API 集成（可选）

### 技术栈

**后端：**
- Go 1.23+
- Beego v2 Web 框架
- XORM 数据库操作
- JWT 令牌管理
- Redis 缓存（可选）

**前端：**
- Vue 3 + TypeScript
- Ant Design Vue 4
- Vite 构建工具
- Pinia 状态管理
- Vue Router 路由管理

### 环境要求

- Go 1.23 或更高版本
- Node.js 18+ 和 pnpm
- PostgreSQL 12 或更高版本
- Redis（可选，用于缓存）

### 快速开始

#### 1. 克隆仓库

```bash
git clone <repository-url>
cd oauth-server
```

#### 2. 后端设置

```bash
# 安装 Go 依赖
go mod download

# 设置 PostgreSQL 数据库
# 详细说明请参见 docs/postgresql-setup.md
createdb oauth_server

# 复制配置文件
copy .env.example .env

# 编辑 .env 并配置：
# - PostgreSQL 连接字符串（必需）
# - JWT 密钥（至少 32 个字符）
# - 管理员凭据（首次运行必需）
# - SMTP 设置（如需邮箱验证）
# - Redis 设置（可选）

# 初始化数据库并创建管理员用户
go run main.go init

# 启动后端服务器
go run main.go
```

后端服务器将在 `http://localhost:8080`（或您配置的端口）启动。

#### 3. 前端设置

```bash
cd frontend

# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev
```

前端将在 `http://localhost:5173` 启动。

生产环境构建：

```bash
pnpm build
```

### 配置说明

编辑 `.env` 进行配置：

**必需设置：**
- 数据库配置（PostgreSQL）：
  - `dbHost` - PostgreSQL 服务器地址（默认：localhost）
  - `dbPort` - PostgreSQL 端口（默认：5432）
  - `dbUser` - 数据库用户名
  - `dbPassword` - 数据库密码
  - `dbName` - 数据库名称
  - `dbSSLMode` - SSL 模式（disable/require/verify-ca/verify-full）
- `adminEmail`、`adminPassword`、`adminUsername` - 初始管理员用户凭据
- `jwtSecret` - JWT 签名密钥（至少 32 个字符）

**数据库配置示例：**
```ini
dbHost = localhost
dbPort = 5432
dbUser = postgres
dbPassword = password
dbName = oauth_server
dbSSLMode = disable
```

详细设置说明请参见 `docs/postgresql-setup.md`

**可选设置：**
- Redis 缓存配置
- SMTP 邮件设置
- CORS 和源站设置
- 实名验证 API

### 项目结构

```
.
├── controllers/       # HTTP 请求处理器
├── models/           # 数据库模型和操作
├── services/         # 业务逻辑层
├── routers/          # 路由定义
├── keys/             # RSA 密钥对（用于加密）
├── frontend/         # Vue 3 前端应用
│   ├── src/
│   │   ├── api/      # API 客户端
│   │   ├── components/  # Vue 组件
│   │   ├── views/    # 页面视图
│   │   ├── router/   # 路由配置
│   │   └── stores/   # Pinia 状态存储
│   └── dist/         # 生产构建输出
└── main.go           # 应用程序入口
```

### API 端点

- `POST /api/auth/login` - 用户登录
- `POST /api/auth/register` - 用户注册
- `POST /api/auth/logout` - 用户登出
- `GET /api/auth/authorize` - OAuth 授权端点
- `POST /api/auth/token` - OAuth 令牌端点
- `GET /api/admin/*` - 管理员管理端点
- `GET /api/user/*` - 用户资料端点

### 默认管理员账户

运行 `go run main.go init` 后，将使用 `.env` 中的凭据创建管理员账户：

- 邮箱：在 `adminEmail` 中配置
- 密码：在 `adminPassword` 中配置
- 用户名：在 `adminUsername` 中配置

**重要提示：** 首次登录后请立即更改这些默认凭据！

### 开发

```bash
# 运行后端（需要 air 或类似工具实现热重载）
go run main.go

# 运行前端开发服务器
cd frontend && pnpm dev

# 构建生产环境前端
cd frontend && pnpm build
```

### 许可证

Apache License 2.0
