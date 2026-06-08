# GoShop

Go 单体商城后端 + uni-app C 端 + Vue3 管理后台。

## 技术栈

- **后端**: Go 1.23+, Gin, GORM, MySQL, Redis, RabbitMQ, Casbin, JWT, Zap, Prometheus
- **C 端**: uni-app Vue3 + Pinia + TypeScript
- **B 端**: Vue3 + Element Plus + Vite
- **部署**: Docker, Kubernetes

## 快速开始（Go 1.23）

### 1. 编译

```bash
go mod tidy
make build
```

生成 `bin/goshop` 与 `bin/seed`。

### 2. 配置数据库

**方式 A — Docker（推荐）**

```bash
make docker-up
# 或: cd deploy && docker compose up -d mysql redis rabbitmq minio
```

默认账号：`goshop` / `goshop123`，库名 `goshop`（见 `config.yaml`）。

**方式 B — 本机已有 MySQL**

```bash
cp config.local.yaml.example config.local.yaml
# 编辑 config.local.yaml，填入你的 user/password
```

并先创建数据库：

```sql
CREATE DATABASE IF NOT EXISTS goshop DEFAULT CHARACTER SET utf8mb4;
```

也可用环境变量覆盖（无需改文件）：

```bash
export GOSHOP_DATABASE_USER=root
export GOSHOP_DATABASE_PASSWORD=你的密码
export GOSHOP_DATABASE_DBNAME=goshop
```

### 3. 初始化种子数据

```bash
make seed
# 或: ./bin/seed -config config.yaml
```

会创建：默认租户、`admin/admin123`、演示商品。

### 4. 启动后端

```bash
make run
# 或: ./bin/goshop -config config.yaml
```

默认 `http://localhost:8888`

### 5. 健康检查

- `GET /healthz` — 存活
- `GET /readyz` — DB 就绪
- `GET /metrics` — Prometheus

### 6. API 联调

```bash
# 管理端登录
curl -X POST http://localhost:8888/api/v1/admin/auth/login \
  -H 'Content-Type: application/json' \
  -H 'X-Tenant-Id: 1' \
  -d '{"username":"admin","password":"admin123"}'

# C端登录
curl -X POST http://localhost:8888/api/v1/app/auth/login \
  -H 'Content-Type: application/json' \
  -H 'X-Tenant-Id: 1' \
  -d '{"phone":"13800138000"}'
```

### 7. 前端

```bash
# 管理端 http://localhost:5174
cd admin-web && npm install && npm run dev

# C端 H5 http://localhost:5173
cd web && npm install && npm run dev:h5
```

## 常用命令

| 命令 | 说明 |
|------|------|
| `make build` | 编译 server + seed |
| `make seed` | 写入种子数据 |
| `make run` | 启动 API 服务 |
| `make docker-up` | Docker 启动中间件 |
| `make dev` | Docker + seed + run 一键脚本 |

## 目录结构

```
goshop/
├── api/openapi.yaml
├── cmd/server/       # API 入口
├── cmd/seed/         # 种子数据
├── internal/
├── web/              # uni-app C端
├── admin-web/        # Vue3 管理端
└── deploy/           # Docker & K8s
```

## K8s 部署

```bash
kubectl apply -f deploy/k8s/
```
