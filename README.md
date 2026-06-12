# Web Clipboard Go

Web Clipboard Go 是一个基于 Go 的临时 Web 剪贴板服务，用于在认证用户之间保存和获取文本、文件。剪贴板条目会在 10 分钟后自动过期。

## 功能特性

- 通过认证 API 保存文本和文件。
- 前端使用 pnpm、Vite 和 React 构建，再由 Go 服务作为静态资源提供。
- 支持最近条目列表，可直接复制文本或下载文件。
- 支持用户登录、退出、密码修改和管理员用户管理。
- 管理/账号功能集中在独立设置页 `/settings.html`。
- 内置文件类型校验、内容检查、访问限流和安全响应头。
- 支持 Docker 和 Docker Compose 部署。

## 项目结构

```text
web-clipboard-go/
├── backend/
│   ├── cmd/web-clipboard/     # Go 应用入口
│   └── internal/              # handlers、middleware、models、services、utils
├── frontend/
│   ├── src/                   # React 组件、认证工具、i18n、Tailwind 入口样式
│   ├── public/                # Vite 构建时复制的静态资源
│   ├── index.html             # 主页面 Vite 入口
│   ├── login.html             # 登录页面 Vite 入口
│   ├── settings.html          # 设置页面 Vite 入口
│   └── package.json           # pnpm 管理的前端依赖和脚本
├── docs/                      # 项目文档
├── .github/workflows/         # GHCR 镜像构建工作流
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── web_contract_test.go       # 前端和路由静态契约测试
```

生成的二进制、运行时数据和本地依赖应放在已忽略目录中，例如 `bin/`、`data/`、`node_modules/`、`frontend/dist/` 和 `.omx/`。

长期持久化数据统一写入数据目录。容器内默认目录是 `/data`，当前用户账号文件为 `/data/users.json`，后续系统设置也应写入同一目录。本地开发如需继续写入仓库下的 `./data`，可设置：

```bash
WEB_CLIPBOARD_DATA_DIR=./data
```

## 构建和运行

本地开发优先使用 Make：

```bash
make build    # 构建 frontend/dist 和 bin/web-clipboard-go.exe
make run      # 构建并在 http://localhost:5000 启动服务
make test     # 构建前端并运行 go test -v ./...
```

也可以手动执行：

```bash
pnpm --dir frontend install --frozen-lockfile
pnpm --dir frontend build
go test ./...
go build -o bin/web-clipboard-go.exe ./backend/cmd/web-clipboard
./bin/web-clipboard-go.exe
```

首次启动时会自动创建默认管理员账号，并在控制台输出初始随机密码：

- Username: `admin`

请妥善保存首次启动日志中的密码，并在首次登录后立即修改。

## API 端点

认证：

- `POST /api/auth/login`
- `POST /api/auth/logout`
- `GET /api/auth/me`

剪贴板：

- `POST /api/text`
- `GET /api/text/{id}`
- `POST /api/file`
- `GET /api/file/{id}`
- `DELETE /api/{id}`
- `GET /api/cleanup`

用户管理：

- `POST /api/users`
- `GET /api/users`
- `GET /api/users/{id}`
- `PUT /api/users/{id}`
- `DELETE /api/users/{id}`
- `PUT /api/users/{id}/password`

## Docker

拉取并运行已发布的 GHCR 镜像：

```bash
docker pull ghcr.io/wang-yn/web-clipboard-go:latest
docker run -d \
  --name web-clipboard-go \
  --restart unless-stopped \
  -p 5000:5000 \
  -v ./data:/data \
  -e GIN_MODE=release \
  ghcr.io/wang-yn/web-clipboard-go:latest
```

用户信息、系统设置等持久化数据统一保存在容器内 `/data` 目录。生产部署时请务必将 `/data` 映射到主机目录，否则删除或重建容器后数据会丢失。

本地镜像开发：

```bash
make docker-build
make docker-run
```

Docker Compose：

```bash
docker compose up -d
docker compose logs -f
docker compose down
```

## Makefile 命令

```bash
make build          # 构建前端和 Go 应用
make frontend-build # 使用 pnpm 和 Vite 构建 frontend/dist
make backend-build  # 只构建 Go 应用
make run            # 本地构建并运行
make test           # 运行测试
make docker-build   # 构建 Docker 镜像
make docker-run     # 运行 Docker 容器
make docker-stop    # 停止并删除 Docker 容器
make docker-logs    # 查看 Docker 容器日志
make compose-up     # 启动 docker-compose 服务
make compose-down   # 停止 docker-compose 服务
make clean          # 清理 Docker 资源和构建产物
make help           # 查看可用命令
```

## 文档规范

- 项目文档统一保存到 `docs/` 目录。
- 简单文档使用 Markdown，文件扩展名为 `.md`。
- 复杂文档使用 HTML，文件扩展名为 `.html`。
- 根目录 `README.md` 只保留项目概览、快速开始、常用命令和关键规范。
- 文档中的代码标识、命令、路径、API 路由、环境变量、包名和配置键保持原样。

## 提交规范

- commit message 只能使用中文。
- 首行写清楚提交意图，说明为什么要改，而不是重复描述改了什么。
- 提交应保持聚焦，避免把无关变更混在同一个 commit 中。
- 可使用 Lore commit protocol 记录关键决策；如果使用 trailer，trailer 的值也必须使用中文。
- trailer 名称保持 git-native 格式，例如 `Constraint:`、`Rejected:`、`Confidence:`、`Scope-risk:`、`Tested:` 和 `Not-tested:`。
- 命令、路径、API 路由、环境变量、包名、配置键和代码符号可保持原样。

## 安全说明

- 不要提交密钥、证书、本地 `.env` 文件、生成的二进制文件或运行时 `data/`。
- 修改认证或用户管理代码时，必须保留文件校验、内容扫描、访问限流、session 过期和最后一个管理员保护逻辑。
- 反向代理必须转发 `/`、`/login.html`、`/settings.html`、`/api/*`、`/assets/*` 和 `/favicon.ico`。
