# WebGUI OIDC 登录

用途：本文说明如何为 Fast Note Sync WebGUI 启用 OpenID Connect (OIDC) 登录。当你希望用户通过 Dex、Keycloak、Casdoor 等外部身份提供方登录 WebGUI 时阅读本文。本文不覆盖 MCP OAuth 资源服务器授权；MCP OAuth 使用 `oauth` 配置。

## 功能目标

`oidc` 配置只用于 WebGUI SSO 登录。

启用后：

- WebGUI 登录页请求 `/api/user/auth/oidc/config`；
- 如果服务端启用了 OIDC，登录页会显示配置的 OIDC 登录按钮；
- `/api/user/auth/oidc/start` 创建 state、nonce 和 PKCE verifier，然后跳转到身份提供方；
- 身份提供方回调到 `oidc.redirect-url`；
- 服务端验证 `id_token`，把 OIDC subject 映射到本地 FNS 用户，并签发正常的 WebGUI 登录 token。

## 配置

在 `config/config.yaml` 中添加 `oidc` 配置：

```yaml
oidc:
  enabled: true
  display-name: "Login with SSO"
  issuer: "https://idp.example.com"
  client-id: "fns-webgui"
  client-secret: "change-me"
  redirect-url: "https://fns.example.com/api/user/auth/oidc/callback"
  callback-path: "/api/user/auth/oidc/callback"
  scopes:
    - openid
    - profile
    - email
  auto-register: false
  user-mapping:
    subject-claim: "sub"
    email-claim: "email"
    username-claim: "preferred_username"
    display-name-claim: "name"
```

当 `enabled: true` 时必须配置：

- `issuer`
- `client-id`
- `client-secret`
- `redirect-url`

默认值：

- `display-name`: `Login with OIDC`
- `callback-path`: `/api/user/auth/oidc/callback`
- `scopes`: `openid`, `profile`, `email`
- `subject-claim`: `sub`
- `email-claim`: `email`
- `username-claim`: `preferred_username`
- `display-name-claim`: `name`

不要把真实的 `client-secret` 提交到公开 Git 配置中。

## 用户映射

FNS 会把 OIDC 绑定关系存储在 `user_oidc_identity` 表中。

登录解析顺序：

1. 如果 `(issuer, subject)` 已经绑定，直接登录对应本地用户。
2. 如果没有绑定，但 OIDC email 匹配已有本地用户，则创建绑定并登录该用户。
3. 如果没有匹配用户且 `auto-register: true`，FNS 会创建本地用户，然后创建绑定。
4. 如果没有匹配用户且 `auto-register: false`，登录失败。

更稳妥的上线方式是先设置 `auto-register: false`，提前创建本地用户，让首次 OIDC 登录通过 email 自动绑定。

当 `auto-register: true` 时，本地用户名会按以下顺序从第一个可用值生成：

1. `username-claim`，例如 `preferred_username`
2. `display-name-claim`，例如 `name`
3. email 中 `@` 前面的部分
4. `oidc_` 加 OIDC subject

生成值会规范化为 FNS 用户名格式：字母、数字、下划线，长度 3 到 20。如果用户名已存在，FNS 会追加数字后缀。

## Provider 配置

### Dex

创建 confidential client：

- Client ID: `fns-webgui`
- Client secret: 与 `oidc.client-secret` 一致
- Redirect URI: `https://fns.example.com/api/user/auth/oidc/callback`
- Scopes: `openid`, `profile`, `email`

`oidc.issuer` 使用 Dex issuer，例如：

```yaml
issuer: "https://dex.example.com/dex"
```

### Keycloak

创建 OpenID Connect confidential client：

- Client ID: `fns-webgui`
- Client authentication: enabled
- Standard flow: enabled
- Valid redirect URI: `https://fns.example.com/api/user/auth/oidc/callback`
- PKCE: 支持 `S256`

`oidc.issuer` 使用 realm issuer：

```yaml
issuer: "https://keycloak.example.com/realms/fns"
```

### Casdoor

创建或更新 application：

- Redirect URI: `https://fns.example.com/api/user/auth/oidc/callback`
- Grant type: `authorization_code`
- Client ID 和 secret 与 `oidc.client-id`、`oidc.client-secret` 一致
- Scopes: `openid`, `profile`, `email`

`oidc.issuer` 使用 Casdoor 对外地址：

```yaml
issuer: "https://casdoor.example.com"
```

Casdoor 常见显示名 claim 是 `displayName`，如需映射可配置：

```yaml
user-mapping:
  display-name-claim: "displayName"
```

## 公网地址与反向代理

`redirect-url` 必须是身份提供方和浏览器都能访问到的外部 callback URL。部署在反向代理后面时，应使用公网 HTTPS 地址，而不是容器内部地址。

示例：

```yaml
redirect-url: "https://notes.example.com/api/user/auth/oidc/callback"
```

如果 WebGUI 使用独立端口，callback 仍属于 API 路由。Provider 中应配置能访问 FNS service 的 callback URL。

## 验证

仓库提供 Docker smoke test：

```bash
scripts/oidc-smoke-test.sh
```

它会在本地启动 Dex、Keycloak、Casdoor，并验证 provider 兼容性。

常规测试不会启动 Docker：

```bash
go test ./...
```

provider smoke test 内部使用 build tag：

```bash
go test -tags oidc_integration ./internal/oidc -run TestOIDCIntegrationProvider
```

## 排错

- `oidc provider discovery failed`：检查 `oidc.issuer` 以及 `/.well-known/openid-configuration`。
- `OIDC state is invalid or expired`：重新开始登录；callback 被重复使用、已过期，或来自另一个服务实例。
- `OIDC token exchange failed`：检查 client ID、client secret、redirect URL 和 PKCE 支持。
- Provider 登录成功但 FNS 登录失败：检查 `email`、`sub` claims，以及是否需要开启 `auto-register`。
- 登录页没有 OIDC 按钮：检查 `oidc.enabled: true`，并确认 WebGUI 能以 `X-Client: WebGui` 请求 `/api/user/auth/oidc/config`。
