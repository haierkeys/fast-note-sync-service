# MCP OAuth 与 Stytch 运行手册

语言：[English](mcp-oauth-stytch.md) | [简体中文](mcp-oauth-stytch.zh-CN.md) | [繁體中文](mcp-oauth-stytch.zh-TW.md)

用途：将 Fast Note Sync Service 配置为受 OAuth 保护的 MCP 资源，供 ChatGPT 和其他支持 OAuth 的 MCP 客户端访问。

当你需要启用 MCP OAuth、把 Stytch 配置为授权服务器、配置 FNS 部署，或者排查 ChatGPT connector 授权问题时，阅读本文。

本文不覆盖 WebGUI OIDC SSO 登录。当前实现不会在 WebGUI 中新增“使用 OIDC 登录”、“使用 Stytch 登录”或“使用 Casdoor 登录”的按钮，也没有实现 `/api/user/auth/callback` 这类 WebGUI 回调路由。

## 当前实现目标

当前 OAuth 实现的目标是 MCP 访问控制。

当 `oauth.enabled` 为 `true` 时，`/api/mcp` 和 `/api/mcp/sse` 会成为支持 OAuth 的受保护资源：

1. 未认证的 MCP 请求会收到 `401 Unauthorized`，响应中包含 `WWW-Authenticate` header。
2. 这个 challenge 会指向 protected resource metadata，通常是 `/.well-known/oauth-protected-resource/api/mcp`。
3. metadata 会声明当前 FNS 部署是受保护资源，并列出配置的授权服务器。
4. ChatGPT 或其他支持 OAuth 的 MCP 客户端会到该授权服务器完成 OAuth。
5. 客户端后续请求 MCP 时会带上 `Authorization: Bearer <access-token>`。
6. FNS 使用配置中的 issuer、JWKS URL、audience/resource 和 scopes 本地验证 bearer JWT。
7. FNS 将验证后的 JWT subject 映射到 FNS 用户，并将 OAuth scopes 映射为 FNS MCP 权限。

这符合 ChatGPT Apps SDK 的认证模型：OAuth 完成后，ChatGPT 会把 access token 附加到 MCP 请求中；MCP server 仍然负责 token 签名验证、issuer/audience 校验、过期时间校验和 scope enforcement。

参考资料：

- OpenAI Apps SDK authentication: <https://developers.openai.com/apps-sdk/build/auth#implementing-token-verification>
- OpenAI Apps SDK MCP overview: <https://developers.openai.com/apps-sdk/concepts/mcp-server#why-apps-sdk-standardises-on-mcp>
- Stytch MCP Server: <https://stytch.com/docs/resources/workspace-management/stytch-mcp-server>
- Stytch custom Connected Apps flow: <https://stytch.com/docs/connected-apps/build-custom-flow/getting-started-api>
- Stytch B2B Start OAuth Authorization API: <https://stytch.com/docs/api-reference/b2b/api/connected-apps/consent-management/start-oauth-authorization>

## 当前不支持什么

这不是 WebGUI OIDC SSO。

当前实现不支持：

- 给 WebGUI 增加 Casdoor 等通用 OIDC provider 的登录按钮。
- WebGUI OIDC callback endpoint，例如 `/api/user/auth/callback`。
- 通过外部 OIDC 登录自动创建或映射 WebGUI 用户。
- 用 Stytch、Casdoor、Pocket ID 或其他 OIDC provider 替换 `/api/user/login`。

`oauth` 配置块用于验证 MCP 请求中的 bearer token，不会改变 WebGUI 登录行为。

## 相关代码路径

- `internal/config/oauth.go`：配置结构、默认值、校验逻辑和 protected resource metadata。
- `internal/routers/router_oauth_metadata.go`：OAuth protected resource metadata 路由。
- `internal/routers/router_mcp.go`：受 OAuth-aware middleware 保护的 MCP 路由。
- `internal/middleware/mcp_oauth.go`：MCP static token 与 OAuth 混合认证。
- `internal/oauth/verifier.go`：使用 issuer、audience、JWKS、过期时间和 scopes 验证 JWT。
- `internal/oauth/subject_mapper.go`：JWT subject 到 FNS UID 的映射。
- `internal/oauth/scope_mapper.go`：OAuth scope 到 FNS MCP 权限的映射。
- `internal/oauth/stytch_client.go`：Stytch authorize start/submit API client。
- `internal/routers/api_router/handler_stytch_oauth.go`：将 authorize page 桥接到 Stytch 的已认证 FNS API。
- `frontend/oauth-authorize.html`：由 WebGUI 构建出的 OAuth authorize page。

## 对外端点

启用 OAuth 后，部署必须通过 HTTPS 暴露这些端点：

| Endpoint | 用途 |
| --- | --- |
| `/api/mcp` | Streamable HTTP MCP endpoint。 |
| `/api/mcp/sse` | 旧版 SSE MCP endpoint。 |
| `/.well-known/oauth-protected-resource/api/mcp` | `/api/mcp` 的 protected resource metadata。 |
| `/.well-known/oauth-protected-resource` | 通用 protected resource metadata endpoint。 |
| `/oauth/authorize` | Stytch Connected App 授权流程使用的浏览器页面。 |
| `/api/oauth/stytch/authorize/start` | 已认证 FNS API，会调用 Stytch authorize start。 |
| `/api/oauth/stytch/authorize/submit` | 已认证 FNS API，会调用 Stytch authorize submit。 |

两个 `/api/oauth/stytch/*` endpoint 都需要已有的 FNS WebGUI login token。如果浏览器中没有 `localStorage.token`，authorize page 会先要求用户登录 FNS。

## 配置参考

在 `config.yaml` 中新增或更新 `oauth` 配置块：

```yaml
oauth:
  enabled: true
  resource: "https://fns.example.com/api/mcp"
  authorization-servers:
    - "https://example.customers.stytch.com"
  jwks-url: "https://example.customers.stytch.com/.well-known/jwks.json"
  issuer: "https://example.customers.stytch.com"
  audience: []
  scopes-supported:
    - notes:read
    - notes:write
    - files:read
    - files:write
    - vaults:read
  required-scopes:
    - notes:read
    - files:read
    - vaults:read
  resource-name: "Fast Note Sync MCP"
  allow-static-fns-token: true
  default-client: "ChatGPT"
  default-client-name: "ChatGPT"
  default-client-version: ""
  default-vault-name: "main"
  default-fns-scope: ""
  subject-mapping:
    mode: "email_or_fixed_uid"
    claim: "email"
    fixed-uid: 1
  stytch:
    enabled: true
    kind: "b2b"
    domain: "https://example.customers.stytch.com"
    project-id: "project-live-..."
    secret: "${FNS_STYTCH_SECRET}"
    organization-id: "organization-live-..."
    member-id: "member-live-..."
```

### 核心 OAuth 字段

| 字段 | 启用后是否必填 | 说明 |
| --- | --- | --- |
| `enabled` | 是 | 启用 MCP OAuth-aware authentication 和 protected resource metadata。 |
| `resource` | 是 | 受保护资源标识。ChatGPT MCP 场景下使用公开 MCP endpoint，例如 `https://fns.example.com/api/mcp`。 |
| `authorization-servers` | 是 | 暴露给 MCP 客户端的授权服务器 origin。Stytch 场景下使用 Stytch customer domain origin。 |
| `jwks-url` | 是 | FNS 用来验证 JWT 签名的 JWKS endpoint。 |
| `issuer` | 是 | 期望的 JWT issuer，必须匹配 token 的 `iss` claim。 |
| `audience` | 否 | 可选的 JWT audience 列表。为空时，FNS 使用 `resource` 作为期望 audience。 |
| `scopes-supported` | 否 | 写入 protected resource metadata 的 scopes。默认是 FNS MCP 标准 scopes。 |
| `required-scopes` | 否 | 每个 OAuth token 必须包含的 scopes。默认是 `notes:read`、`files:read` 和 `vaults:read`。 |
| `resource-name` | 否 | metadata 中展示的人类可读资源名。 |
| `allow-static-fns-token` | 否 | 启用 OAuth 时是否继续允许原有 FNS static token MCP 客户端访问。默认 true。 |

只有在授权服务器或客户端无法签发 FNS 自定义 scopes，并且你明确希望依赖 FNS permission mapping 或部署层控制时，才使用 `required-scopes: []`。更严格的公网部署应保留显式 required scopes。

### FNS 身份与权限映射

| 字段 | 说明 |
| --- | --- |
| `subject-mapping.mode` | `email`、`fixed_uid` 或 `email_or_fixed_uid`。 |
| `subject-mapping.claim` | 通过邮箱匹配 FNS 用户时读取的 JWT claim，默认 `email`。 |
| `subject-mapping.fixed-uid` | `fixed_uid` 模式使用的 FNS UID，或 `email_or_fixed_uid` 模式的 fallback UID。 |
| `default-fns-scope` | 如果设置，则跳过 OAuth scope mapping，直接给已验证 OAuth 请求授予该 FNS scope。 |
| `default-client`, `default-client-name`, `default-client-version` | 请求或 token 没提供客户端信息时使用的默认 MCP client headers。 |
| `default-vault-name` | MCP 操作未提供 vault 参数时使用的默认笔记库名称。 |

单用户部署最简单的配置是 `fixed_uid`：

```yaml
subject-mapping:
  mode: fixed_uid
  fixed-uid: 1
```

多用户部署建议使用 email mapping：

```yaml
subject-mapping:
  mode: email
  claim: email
```

被验证的 access token 中必须存在对应 email claim，并且 FNS 中必须已经存在该用户。

## Stytch 配置

使用 Stytch Connected Apps 作为授权服务器。FNS server 自身不是完整 OAuth authorization server；它是 MCP protected resource，并使用 Stytch 签发 token。

### 1. 创建或选择 Stytch project

可选类型：

- B2B project：推荐用于受控的 organization/member 访问。
- Consumer project：适用于不需要 organization/member 标识的用户访问。

记录这些值：Stytch project ID、Stytch project secret、Stytch customer domain，以及 B2B 场景下的 organization ID 和 member ID。

project secret 只能存放在 secret manager、环境变量或其他私有配置机制中，不要提交真实 secret 到 Git。

### 2. 配置 Connected App

在 Stytch 中为 ChatGPT 或目标 MCP 客户端创建 Connected App。

需要配置：

- Redirect URI：MCP 客户端或 ChatGPT connector setup 要求的 redirect URI。
- Allowed scopes：加入客户端会请求的 FNS MCP scopes，例如 `notes:read`、`notes:write`、`files:read`、`files:write`、`vaults:read`。
- Authorization entry point：指向 FNS authorize page，例如 `https://fns.example.com/oauth/authorize`。

authorize page 会解析 `client_id`、`redirect_uri`、`response_type`、`scope`、`state`、`nonce`、`code_challenge`、`code_challenge_method` 等 OAuth 参数，然后调用已认证的 FNS Stytch bridge endpoints。

### 可选：通过 Stytch MCP 配置

Stytch 也提供自己的 MCP server：

```text
https://mcp.stytch.dev/mcp
```

Stytch MCP 可以在支持 MCP 的开发环境中用自然语言管理 Stytch project 配置，而不必全部通过 Dashboard 点击完成。它仍然需要先完成 Stytch workspace 的 OAuth 授权，才能读取或修改 workspace 资源。

可用时，可以用 Stytch MCP 检查或配置 project 类型、Connected App client、redirect URIs、public tokens、SDK settings，以及 Connected Apps 需要的 workspace-level settings。

即使使用 Stytch MCP，FNS 的 `config.yaml` 仍然是 FNS 部署配置的来源。Stytch MCP 只配置 Stytch，不会更新 FNS 服务配置。

### 3. 将 Stytch 值写入 FNS

B2B 示例：

```yaml
oauth:
  stytch:
    enabled: true
    kind: "b2b"
    domain: "https://example.customers.stytch.com"
    project-id: "project-live-..."
    secret: "${FNS_STYTCH_SECRET}"
    organization-id: "organization-live-..."
    member-id: "member-live-..."
```

Consumer 示例：

```yaml
oauth:
  stytch:
    enabled: true
    kind: "consumer"
    domain: "https://example.customers.stytch.com"
    project-id: "project-live-..."
    secret: "${FNS_STYTCH_SECRET}"
    user-id-prefix: "fns:"
```

Consumer mode 会优先使用 `stytch.user-id`。如果没有配置，则发送 `stytch.user-id-prefix + <FNS UID>` 给 Stytch。

## 启用步骤

1. 通过 HTTPS 部署 FNS，确保公网 base URL 稳定，例如 `https://fns.example.com`。
2. 设置 `server.ext-api-url` 为公网 origin。
3. 配置 Stytch Connected App，并收集 domain、project ID、secret，以及 B2B 场景下的 organization/member IDs。
4. 添加 `oauth` 配置块。`resource` 设置为 `https://fns.example.com/api/mcp`，`issuer`、`jwks-url` 和 `authorization-servers` 设置为 Stytch customer domain 及其 JWKS URL。
5. 安全存储 Stytch secret。使用你的常规 secret 管理方式，或使用未提交的本地配置文件，不要提交真实 Stytch secret。
6. 重启 FNS 服务。
7. 验证 protected resource metadata：

   ```bash
   curl -sS https://fns.example.com/.well-known/oauth-protected-resource/api/mcp
   ```

8. 验证 MCP challenge：

   ```bash
   curl -i https://fns.example.com/api/mcp
   ```

   期望返回 `401`，并包含指向 metadata 的 `WWW-Authenticate` header。

9. 在 ChatGPT 或其他支持 OAuth 的 MCP 客户端中添加 MCP server：`https://fns.example.com/api/mcp`。
10. 授权过程中，如果 authorize page 要求登录 FNS，请先登录 FNS，因为 bridge endpoints 使用现有 FNS WebGUI token 保护。
11. consent 完成后，确认 MCP 客户端可以列出并调用 tools。

## Static token 兼容

如果 `allow-static-fns-token: true`，现有 MCP 客户端仍可继续使用 FNS static token：

```http
Authorization: Bearer <fns-token>
```

如果需要 OAuth-only MCP access：

```yaml
oauth:
  allow-static-fns-token: false
```

## Troubleshooting

### ChatGPT 提示 MCP server does not implement OAuth

检查：

- `oauth.enabled` 为 true。
- `https://fns.example.com/.well-known/oauth-protected-resource/api/mcp` 返回 HTTP 200。
- `https://fns.example.com/api/mcp` 返回 `401`，并且 `WWW-Authenticate` header 包含 `resource_metadata`。
- `oauth.resource` 与公开 MCP endpoint URL 完全一致。
- ingress 暴露了 `/.well-known/oauth-protected-resource/api/mcp`。

### Metadata 缺少 `authorization_servers`

检查 `oauth.authorization-servers`。它必须至少包含一个授权服务器 origin。

### Token verification 返回 `invalid_token`

检查 token 的 `iss` claim、`oauth.jwks-url` 可达性、JWKS 中是否存在签名 key、token 是否过期，以及 `oauth.audience` 或 `oauth.resource` 是否匹配 token audience/resource。

### Token verification 返回 `insufficient_scope`

检查 token 是否包含 `oauth.required-scopes` 中列出的 scopes、Stytch Connected App 是否允许签发请求的 FNS scopes，以及 `default-fns-scope` 是否符合预期。

### WebGUI 没有 OIDC 登录按钮

这是预期行为。当前实现不提供 WebGUI OIDC SSO，只保护 MCP 路由，并为 MCP 客户端提供 OAuth authorization bridge。

### `/api/user/auth/callback` 返回 404

这是预期行为。该 endpoint 没有实现。

### Stytch authorize start 或 submit 失败

检查 `oauth.stytch.enabled`、`oauth.stytch.domain`、`project-id`、`secret`、B2B 的 `organization-id` 和 `member-id`、Connected App 的 `client_id`/`redirect_uri`/scopes，以及用户是否已在 `/oauth/authorize` 前登录 FNS。
