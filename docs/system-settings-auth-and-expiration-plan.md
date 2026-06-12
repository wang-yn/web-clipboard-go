# 设置页登录方式与剪贴板有效期方案

## 目标

在 `/settings.html` 的系统设置区域增加可持久化配置：

- 用户密码登录开关。
- Google 登录开关及必要配置。
- GitHub 登录开关及必要配置。
- 剪贴板有效时间设置，默认 `10 分钟`，单位支持分钟、小时、天、永不过期。

该能力应保持现有本地用户、角色、session 和 OAuth 回调模型不变，只把登录入口和剪贴板有效期从硬编码/环境变量升级为可由管理员配置的系统设置。

## 当前现状

- `/settings.html` 已存在，并有系统设置占位。
- 用户数据和后续长期配置应统一写入 `/data`。
- OAuth provider 当前由环境变量初始化。
- `POST /api/auth/login` 始终允许本地用户名密码登录。
- 文本和文件保存时都硬编码 `time.Now().UTC().Add(10 * time.Minute)`。
- 过期判断直接使用 `item.ExpiresAt.Before(now)`，没有永不过期语义。

## 推荐设计

### 后端设置服务

新增 `SettingsService`，持久化文件：

```text
/data/settings.json
```

默认配置：

```json
{
  "auth": {
    "passwordLoginEnabled": true,
    "oauthAutoProvision": false,
    "allowedEmailDomains": [],
    "google": {
      "enabled": false,
      "clientId": "",
      "clientSecret": ""
    },
    "github": {
      "enabled": false,
      "clientId": "",
      "clientSecret": ""
    }
  },
  "clipboard": {
    "expirationValue": 10,
    "expirationUnit": "minute"
  }
}
```

建议模型：

```go
type SystemSettings struct {
    Auth      AuthSettings      `json:"auth"`
    Clipboard ClipboardSettings `json:"clipboard"`
}

type AuthSettings struct {
    PasswordLoginEnabled bool                `json:"passwordLoginEnabled"`
    OAuthAutoProvision   bool                `json:"oauthAutoProvision"`
    AllowedEmailDomains  []string            `json:"allowedEmailDomains"`
    Google               OAuthProviderConfig `json:"google"`
    GitHub               OAuthProviderConfig `json:"github"`
}

type OAuthProviderConfig struct {
    Enabled      bool   `json:"enabled"`
    ClientID     string `json:"clientId"`
    ClientSecret string `json:"clientSecret"`
}

type ClipboardSettings struct {
    ExpirationValue int    `json:"expirationValue"`
    ExpirationUnit  string `json:"expirationUnit"`
}
```

### API

新增管理员接口：

```text
GET /api/settings
PUT /api/settings
```

规则：

- 仅管理员可读取和修改系统设置。
- `GET /api/settings` 不返回明文 `clientSecret`，只返回 `clientSecretSet`。
- `PUT /api/settings` 中 `clientSecret` 为空时保留旧值；显式 `clearClientSecret=true` 时清空。
- 保存时必须保证至少一种登录方式可用。
- provider 开启但 `clientId` 或 `clientSecret` 缺失时不视为可用登录方式。

### 登录设置行为

`POST /api/auth/login`：

- `passwordLoginEnabled=true` 时保持现有行为。
- `passwordLoginEnabled=false` 时返回 `403`，不再校验密码。

`GET /api/auth/providers`：

- 返回已启用且配置完整的 Google/GitHub provider。
- 附带 `passwordLoginEnabled`，让登录页知道是否展示本地密码表单。

OAuth provider 创建：

- `OAuthService` 从 `SettingsService` 动态读取配置。
- `OAUTH_AUTO_PROVISION` 和 `OAUTH_ALLOWED_EMAIL_DOMAINS` 改由 settings 管理。
- 环境变量可作为首次默认配置来源，但运行期以 `/data/settings.json` 为准。

### 剪贴板有效期行为

新增统一过期策略：

```go
func (s SystemSettings) ClipboardExpiresAt(now time.Time) time.Time
func ClipboardItemExpired(item *ClipboardItem, now time.Time) bool
```

单位映射：

- `minute`: `value * time.Minute`
- `hour`: `value * time.Hour`
- `day`: `value * 24 * time.Hour`
- `never`: 返回 `time.Time{}` 表示永不过期

所有过期判断使用 helper：

- `GetText`
- `GetFile`
- `ListRecentItems`
- `Cleanup`
- `performCleanup`

前端展示：

- 永不过期条目的 `expiresAt` 可为空或零值；前端应显示“永不过期”。

### 前端设置页

将 `frontend/src/settings.jsx` 的系统设置占位替换为真实表单。

系统设置区域包含：

- 用户密码登录开关。
- 自动创建 OAuth 用户开关。
- 允许邮箱域名输入框，逗号分隔。
- Google 登录开关、Client ID、Client Secret。
- GitHub 登录开关、Client ID、Client Secret。
- 剪贴板有效时间数字输入。
- 单位下拉：分钟、小时、天、永不过期。
- 保存按钮。

交互规则：

- 非管理员不显示系统设置表单。
- 选择永不过期时禁用数字输入。
- secret 已存在时显示“已配置”，不回显明文。
- 保存成功后显示 toast。

## 测试计划

后端单元测试：

- 默认设置为密码登录开启、剪贴板 `10 minute`。
- settings 保存后可重新读取。
- `GET /api/settings` 不返回明文 secret。
- 禁止保存没有任何可用登录方式的配置。
- 密码登录关闭后 `/api/auth/login` 返回 `403`。
- provider 只有开启且配置完整才出现在 `/api/auth/providers`。
- 剪贴板过期策略支持分钟、小时、天。
- 永不过期条目不会被读取、列表、cleanup 判定为过期。

前端契约测试：

- 设置页包含系统设置表单入口。
- 包含 password/google/github 开关。
- 包含有效时间数字输入和单位选择。
- `Auth` 提供 settings 读取和保存方法。

最终验证：

```bash
make test
```
