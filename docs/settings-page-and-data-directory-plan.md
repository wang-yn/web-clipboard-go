# 设置页拆分与数据目录统一方案

## 目标

将主页面中的管理/设置类功能迁移到独立设置页面，并统一所有长期持久化数据的写入位置。用户信息、系统设置等持久化数据统一写入容器内 `/data` 目录，Docker 部署文档明确提示用户将 `/data` 映射到主机目录，避免容器重建后数据丢失。

## 当前现状

- 主页面 `frontend/src/App.jsx` 同时承载剪贴板核心功能、账号操作和管理员用户管理。
- `UserManagement` 当前嵌入主页面，仅管理员可见。
- `ChangePasswordModal` 当前由主页面顶部账号区域触发。
- 前端当前是 Vite 多入口结构，已有 `index.html` 和 `login.html`。
- 后端 Gin 当前只服务 `/` 和 `/login.html` 两个页面入口。
- 用户数据当前通过 `NewUserManager("./data")` 写入工作目录下的 `./data/users.json`。
- Dockerfile 当前没有创建独立 `/data` 目录。
- `docker-compose.yml` 当前没有默认挂载持久化数据目录。
- README 的 Docker 示例没有提示挂载 `/data`。

## 范围

### 设置页拆分

新增独立设置页面：

- `frontend/settings.html`
- `frontend/src/settings.jsx`

主页面保留：

- 文本剪贴板
- 文件剪贴板
- 最近项目
- 当前用户信息
- 语言切换
- 退出登录
- 设置入口按钮

设置页承载：

- 当前账号信息
- 修改当前用户密码
- 管理员用户管理
- 后续系统设置入口

### 数据目录统一

所有长期持久化数据统一写入：

```text
/data
```

当前至少包括：

```text
/data/users.json
```

后续系统设置建议使用：

```text
/data/settings.json
```

临时文件仍可继续使用系统临时目录，例如：

```text
/tmp/web-clipboard-go
```

临时文件不属于长期持久化数据。

## 推荐设计

### 1. 前端页面结构

沿用当前 Vite 多页面模式，不引入 React Router。

推荐新增：

```text
frontend/settings.html
frontend/src/settings.jsx
```

更新：

```text
frontend/vite.config.js
```

将 `settings.html` 加入 Vite build input：

```js
input: {
  index: resolve(__dirname, 'index.html'),
  login: resolve(__dirname, 'login.html'),
  settings: resolve(__dirname, 'settings.html')
}
```

### 2. 前端组件拆分

建议从 `App.jsx` 中抽出共享组件，降低主页面和设置页之间的耦合。

推荐结构：

```text
frontend/src/App.jsx
frontend/src/settings.jsx
frontend/src/shared.jsx
frontend/src/account.jsx
frontend/src/users.jsx
```

职责划分：

```text
shared.jsx
- IconLabel
- StatusMessage
- Modal
- ModalActions
- TextField
- PasswordField

account.jsx
- AccountMenu
- ChangePasswordModal

users.jsx
- UserManagement
- UserFormModal
- ResetPasswordModal

App.jsx
- AppShell
- ClipboardPanel
- RecentItems

settings.jsx
- SettingsShell
- AccountSettings
- AdminSettings
```

### 3. 主页面调整

主页面移除：

```text
user?.role === 'admin' && <UserManagement />
```

主页面新增设置入口：

```text
/settings.html
```

推荐仅管理员显示“用户管理”不等于仅管理员可进设置页。普通用户也应能进入设置页修改自己的密码。

### 4. 设置页访问逻辑

设置页启动时执行认证检查：

```text
Auth.requireAuth()
```

认证失败跳转：

```text
/login.html
```

普通用户可见：

- 当前账号信息
- 修改密码
- 返回剪贴板

管理员额外可见：

- 用户管理
- 创建用户
- 编辑用户
- 启停用户
- 删除用户
- 重置用户密码

普通用户不应触发 `/api/users` 列表请求，避免无意义的 403。

### 5. 后端页面路由

新增 Gin 静态页面路由：

```go
router.GET("/settings.html", func(c *gin.Context) {
    c.File("./frontend/dist/settings.html")
})
```

保留：

```go
router.GET("/login.html", ...)
router.GET("/", ...)
```

API 权限边界无需大改：

- `/api/users` 系列继续管理员专用。
- `/api/users/:id/password` 继续保持 authenticated but not admin-only。
- 普通用户只能修改自己的密码。
- 管理员可以重置其他用户密码。

### 6. 数据目录统一

新增统一数据目录解析函数：

```go
func getDataDir() string {
    if value := os.Getenv("WEB_CLIPBOARD_DATA_DIR"); value != "" {
        return value
    }
    return "/data"
}
```

启动逻辑从：

```go
services.NewUserManager("./data")
```

调整为：

```go
services.NewUserManager(getDataDir())
```

这样容器内默认写入：

```text
/data/users.json
```

本地开发如需保留当前行为，可通过环境变量指定：

```powershell
$env:WEB_CLIPBOARD_DATA_DIR = "./data"
```

或在 Makefile 中显式设置本地开发默认值。

### 7. 系统设置持久化预留

如果本次设置页只迁移用户管理和密码修改，可以先不实现系统设置 API。

但如果要支持系统设置，应新增：

```text
backend/internal/services/settings.go
```

数据文件：

```text
/data/settings.json
```

建议模型：

```go
type SettingsData struct {
    Settings SystemSettings `json:"settings"`
}

type SystemSettings struct {
    // 后续按实际需求增加字段
}
```

原则：

- 不把系统设置写入前端 localStorage。
- 不写入 `/app` 工作目录。
- 不写入临时目录。
- 所有长期配置都归档到 `/data`。

### 8. Dockerfile 调整

Dockerfile final stage 中创建 `/data` 并授权给非 root 用户：

```dockerfile
RUN mkdir -p /data && chown -R appuser:appuser /app /data

VOLUME ["/data"]
```

如果当前已有：

```dockerfile
RUN chown -R appuser:appuser /app
```

可调整为：

```dockerfile
RUN mkdir -p /data && chown -R appuser:appuser /app /data
```

### 9. docker-compose.yml 调整

默认挂载持久化数据目录：

```yaml
services:
  web-clipboard-go:
    volumes:
      - ./data:/data
```

原来的临时目录挂载说明可以删除或降级为高级说明：

```yaml
# Optional: mount temp files only if you need to inspect temporary uploads.
# - ./temp:/tmp/web-clipboard-go
```

### 10. README 调整

Docker run 示例增加：

```bash
-v ./data:/data
```

示例：

```bash
docker run -d \
  --name web-clipboard-go \
  --restart unless-stopped \
  -p 5000:5000 \
  -v ./data:/data \
  -e GIN_MODE=release \
  ghcr.io/wang-yn/web-clipboard-go:latest
```

新增说明：

```text
Persistent data is stored under /data inside the container.
Mount this directory to a host path before production use.
It contains user accounts and future system settings.
```

中文说明可写为：

```text
用户信息、系统设置等持久化数据统一保存在容器内 /data 目录。
生产部署时请务必将 /data 映射到主机目录，否则删除或重建容器后数据会丢失。
```

反向代理说明增加：

```text
/settings.html
```

最终代理路径至少包括：

```text
/
/login.html
/settings.html
/api/*
/assets/*
/favicon.ico
```

### 11. 测试调整

需要更新 `web_contract_test.go`。

当前测试硬编码要求 `UserManagement` 存在于 `frontend/src/App.jsx`。拆分后应改为检查功能存在于设置页相关源码中。

建议新增或调整测试：

- `settings.html` 存在 React root 和 module script。
- `vite.config.js` 包含 `settings.html` input。
- Gin router 服务 `settings.html`。
- 主页面包含 `/settings.html` 设置入口。
- `App.jsx` 不再直接渲染 `UserManagement`。
- 设置页源码包含 `UserManagement`。
- 普通用户改密流程仍使用 `/api/users/${userId}/password`。
- `/api/users/:id/password` 仍不是 admin-only。
- Dockerfile 创建 `/data`。
- `docker-compose.yml` 挂载 `./data:/data`。
- README Docker 示例包含 `/data` 挂载说明。

## 实施步骤

1. 抽取共享前端组件。
2. 新增设置页 HTML 和 React 入口。
3. 将 `UserManagement` 和改密入口迁移到设置页。
4. 主页面增加设置入口，并移除管理员管理区域。
5. 更新 Vite 多入口配置。
6. 后端新增 `/settings.html` 静态路由。
7. 新增统一数据目录函数，默认 `/data`。
8. 用户数据路径改为 `/data/users.json`。
9. Dockerfile 创建并授权 `/data`。
10. docker-compose.yml 默认挂载 `./data:/data`。
11. README 更新 Docker 部署、数据持久化和反向代理说明。
12. 更新静态合同测试。
13. 运行验证命令。

## 验收标准

- 登录后主页面只展示剪贴板核心功能和设置入口。
- 管理员用户管理不再出现在主页面。
- `/settings.html` 可以直接访问。
- 未登录访问设置页会跳转 `/login.html`。
- 普通用户访问设置页只能看到账号设置。
- 管理员访问设置页可以看到用户管理。
- 修改当前密码后仍会清理认证并要求重新登录。
- 管理员重置他人密码功能保持可用。
- 用户数据写入 `/data/users.json`。
- Docker 容器使用 `-v ./data:/data` 后，重建容器不丢失用户数据。
- `make test` 通过。

## 验证命令

```bash
pnpm --dir frontend build
go test ./...
make test
```

可选手工验证：

```text
1. 启动服务。
2. 使用首次启动控制台输出的 admin 初始随机密码登录。
3. 确认主页面有设置入口。
4. 进入 /settings.html。
5. 创建普通用户。
6. 重启容器。
7. 确认普通用户仍存在。
```

## 风险与缓解

| 风险 | 影响 | 缓解 |
| --- | --- | --- |
| 组件拆分导致引用遗漏 | 前端构建失败 | 先抽共享组件，再迁移页面 |
| 普通用户访问设置页触发管理员 API | 出现 403 或错误提示 | 按 `user.role === 'admin'` 延迟渲染 `UserManagement` |
| Docker 用户无权限写 `/data` | 用户创建或改密失败 | Dockerfile 中 `chown appuser:appuser /data` |
| 本地开发默认写入 `/data` 不方便 | Windows 本地运行路径异常 | 支持 `WEB_CLIPBOARD_DATA_DIR` 覆盖 |
| 测试仍绑定旧文件位置 | 合同测试误报失败 | 将测试从“组件在 App.jsx”改为“功能存在且页面入口正确” |

## 推荐提交拆分

建议拆成两个提交：

1. `拆分设置页以隔离账号和管理员功能`
2. `统一持久化数据目录并更新 Docker 部署说明`

如果希望一次提交，也可以合并为：

```text
隔离设置页并固定持久化数据目录
```
