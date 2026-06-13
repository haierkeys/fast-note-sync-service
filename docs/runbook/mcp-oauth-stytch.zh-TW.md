# MCP OAuth 與 Stytch 運行手冊

語言：[English](mcp-oauth-stytch.md) | [简体中文](mcp-oauth-stytch.zh-CN.md) | [繁體中文](mcp-oauth-stytch.zh-TW.md)

用途：將 Fast Note Sync Service 設定為受 OAuth 保護的 MCP 資源，供 ChatGPT 與其他支援 OAuth 的 MCP 用戶端存取。

當你需要啟用 MCP OAuth、把 Stytch 設定為授權伺服器、設定 FNS 部署，或排查 ChatGPT connector 授權問題時，閱讀本文。

本文不涵蓋 WebGUI OIDC SSO 登入。當前實作不會在 WebGUI 中新增「使用 OIDC 登入」、「使用 Stytch 登入」或「使用 Casdoor 登入」按鈕，也沒有實作 `/api/user/auth/callback` 這類 WebGUI callback route。

## 當前實作目標

當前 OAuth 實作的目標是 MCP 存取控制。

當 `oauth.enabled` 為 `true` 時，`/api/mcp` 與 `/api/mcp/sse` 會成為支援 OAuth 的受保護資源：

1. 未認證的 MCP request 會收到 `401 Unauthorized`，response 中包含 `WWW-Authenticate` header。
2. 這個 challenge 會指向 protected resource metadata，通常是 `/.well-known/oauth-protected-resource/api/mcp`。
3. metadata 會宣告目前 FNS 部署是受保護資源，並列出設定的授權伺服器。
4. ChatGPT 或其他支援 OAuth 的 MCP 用戶端會到該授權伺服器完成 OAuth。
5. 用戶端後續請求 MCP 時會帶上 `Authorization: Bearer <access-token>`。
6. FNS 使用設定中的 issuer、JWKS URL、audience/resource 與 scopes 本地驗證 bearer JWT。
7. FNS 將驗證後的 JWT subject 映射到 FNS 使用者，並將 OAuth scopes 映射為 FNS MCP 權限。

這符合 ChatGPT Apps SDK 的認證模型：OAuth 完成後，ChatGPT 會把 access token 附加到 MCP requests；MCP server 仍然負責 token 簽名驗證、issuer/audience 檢查、過期時間檢查與 scope enforcement。

參考資料：

- OpenAI Apps SDK authentication: <https://developers.openai.com/apps-sdk/build/auth#implementing-token-verification>
- OpenAI Apps SDK MCP overview: <https://developers.openai.com/apps-sdk/concepts/mcp-server#why-apps-sdk-standardises-on-mcp>
- Stytch MCP Server: <https://stytch.com/docs/resources/workspace-management/stytch-mcp-server>
- Stytch custom Connected Apps flow: <https://stytch.com/docs/connected-apps/build-custom-flow/getting-started-api>
- Stytch B2B Start OAuth Authorization API: <https://stytch.com/docs/api-reference/b2b/api/connected-apps/consent-management/start-oauth-authorization>

## 當前不支援什麼

這不是 WebGUI OIDC SSO。

當前實作不支援：

- 給 WebGUI 新增 Casdoor 等通用 OIDC provider 的登入按鈕。
- WebGUI OIDC callback endpoint，例如 `/api/user/auth/callback`。
- 透過外部 OIDC 登入自動建立或映射 WebGUI 使用者。
- 用 Stytch、Casdoor、Pocket ID 或其他 OIDC provider 取代 `/api/user/login`。

`oauth` 設定區塊用於驗證 MCP request 中的 bearer token，不會改變 WebGUI 登入行為。

## 相關程式碼路徑

- `internal/config/oauth.go`：設定結構、預設值、校驗邏輯與 protected resource metadata。
- `internal/routers/router_oauth_metadata.go`：OAuth protected resource metadata routes。
- `internal/routers/router_mcp.go`：受 OAuth-aware middleware 保護的 MCP routes。
- `internal/middleware/mcp_oauth.go`：MCP static token 與 OAuth 混合認證。
- `internal/oauth/verifier.go`：使用 issuer、audience、JWKS、過期時間與 scopes 驗證 JWT。
- `internal/oauth/subject_mapper.go`：JWT subject 到 FNS UID 的映射。
- `internal/oauth/scope_mapper.go`：OAuth scope 到 FNS MCP 權限的映射。
- `internal/oauth/stytch_client.go`：Stytch authorize start/submit API client。
- `internal/routers/api_router/handler_stytch_oauth.go`：將 authorize page 橋接到 Stytch 的已認證 FNS API。
- `frontend/oauth-authorize.html`：由 WebGUI 建置出的 OAuth authorize page。

## 對外端點

啟用 OAuth 後，部署必須透過 HTTPS 暴露這些端點：

| Endpoint | 用途 |
| --- | --- |
| `/api/mcp` | Streamable HTTP MCP endpoint。 |
| `/api/mcp/sse` | 舊版 SSE MCP endpoint。 |
| `/.well-known/oauth-protected-resource/api/mcp` | `/api/mcp` 的 protected resource metadata。 |
| `/.well-known/oauth-protected-resource` | 通用 protected resource metadata endpoint。 |
| `/oauth/authorize` | Stytch Connected App 授權流程使用的瀏覽器頁面。 |
| `/api/oauth/stytch/authorize/start` | 已認證 FNS API，會呼叫 Stytch authorize start。 |
| `/api/oauth/stytch/authorize/submit` | 已認證 FNS API，會呼叫 Stytch authorize submit。 |

兩個 `/api/oauth/stytch/*` endpoint 都需要已有的 FNS WebGUI login token。如果瀏覽器中沒有 `localStorage.token`，authorize page 會先要求使用者登入 FNS。

## 設定參考

在 `config.yaml` 中新增或更新 `oauth` 設定區塊：

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

### 核心 OAuth 欄位

| 欄位 | 啟用後是否必填 | 說明 |
| --- | --- | --- |
| `enabled` | 是 | 啟用 MCP OAuth-aware authentication 與 protected resource metadata。 |
| `resource` | 是 | 受保護資源識別。ChatGPT MCP 情境下使用公開 MCP endpoint，例如 `https://fns.example.com/api/mcp`。 |
| `authorization-servers` | 是 | 暴露給 MCP 用戶端的授權伺服器 origin。Stytch 情境下使用 Stytch customer domain origin。 |
| `jwks-url` | 是 | FNS 用來驗證 JWT 簽名的 JWKS endpoint。 |
| `issuer` | 是 | 預期的 JWT issuer，必須符合 token 的 `iss` claim。 |
| `audience` | 否 | 可選的 JWT audience 清單。為空時，FNS 使用 `resource` 作為預期 audience。 |
| `scopes-supported` | 否 | 寫入 protected resource metadata 的 scopes。預設是 FNS MCP 標準 scopes。 |
| `required-scopes` | 否 | 每個 OAuth token 必須包含的 scopes。預設是 `notes:read`、`files:read` 與 `vaults:read`。 |
| `resource-name` | 否 | metadata 中展示的人類可讀資源名。 |
| `allow-static-fns-token` | 否 | 啟用 OAuth 時是否繼續允許原有 FNS static token MCP 用戶端存取。預設 true。 |

只有在授權伺服器或用戶端無法簽發 FNS 自訂 scopes，且你明確希望依賴 FNS permission mapping 或部署層控制時，才使用 `required-scopes: []`。更嚴格的公開部署應保留明確 required scopes。

### FNS 身分與權限映射

| 欄位 | 說明 |
| --- | --- |
| `subject-mapping.mode` | `email`、`fixed_uid` 或 `email_or_fixed_uid`。 |
| `subject-mapping.claim` | 透過信箱匹配 FNS 使用者時讀取的 JWT claim，預設 `email`。 |
| `subject-mapping.fixed-uid` | `fixed_uid` 模式使用的 FNS UID，或 `email_or_fixed_uid` 模式的 fallback UID。 |
| `default-fns-scope` | 如果設定，則跳過 OAuth scope mapping，直接給已驗證 OAuth request 授予該 FNS scope。 |
| `default-client`, `default-client-name`, `default-client-version` | request 或 token 未提供用戶端資訊時使用的預設 MCP client headers。 |
| `default-vault-name` | MCP 操作未提供 vault 參數時使用的預設筆記庫名稱。 |

單使用者部署最簡單的設定是 `fixed_uid`：

```yaml
subject-mapping:
  mode: fixed_uid
  fixed-uid: 1
```

多使用者部署建議使用 email mapping：

```yaml
subject-mapping:
  mode: email
  claim: email
```

被驗證的 access token 中必須存在對應 email claim，且 FNS 中必須已經存在該使用者。

## Stytch 設定

使用 Stytch Connected Apps 作為授權伺服器。FNS server 本身不是完整 OAuth authorization server；它是 MCP protected resource，並使用 Stytch 簽發 token。

### 1. 建立或選擇 Stytch project

可選類型：

- B2B project：建議用於受控的 organization/member 存取。
- Consumer project：適用於不需要 organization/member 識別的使用者存取。

記錄這些值：Stytch project ID、Stytch project secret、Stytch customer domain，以及 B2B 情境下的 organization ID 與 member ID。

project secret 只能存放在 secret manager、環境變數或其他私有設定機制中，不要提交真實 secret 到 Git。

### 2. 設定 Connected App

在 Stytch 中為 ChatGPT 或目標 MCP 用戶端建立 Connected App。

需要設定：

- Redirect URI：MCP 用戶端或 ChatGPT connector setup 要求的 redirect URI。
- Allowed scopes：加入用戶端會請求的 FNS MCP scopes，例如 `notes:read`、`notes:write`、`files:read`、`files:write`、`vaults:read`。
- Authorization entry point：指向 FNS authorize page，例如 `https://fns.example.com/oauth/authorize`。

authorize page 會解析 `client_id`、`redirect_uri`、`response_type`、`scope`、`state`、`nonce`、`code_challenge`、`code_challenge_method` 等 OAuth 參數，然後呼叫已認證的 FNS Stytch bridge endpoints。

### 可選：透過 Stytch MCP 設定

Stytch 也提供自己的 MCP server：

```text
https://mcp.stytch.dev/mcp
```

Stytch MCP 可以在支援 MCP 的開發環境中用自然語言管理 Stytch project 設定，而不必全部透過 Dashboard 點擊完成。它仍然需要先完成 Stytch workspace 的 OAuth 授權，才能讀取或修改 workspace 資源。

可用時，可以用 Stytch MCP 檢查或設定 project 類型、Connected App client、redirect URIs、public tokens、SDK settings，以及 Connected Apps 需要的 workspace-level settings。

即使使用 Stytch MCP，FNS 的 `config.yaml` 仍然是 FNS 部署設定的來源。Stytch MCP 只設定 Stytch，不會更新 FNS 服務設定。

### 3. 將 Stytch 值寫入 FNS

B2B 範例：

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

Consumer 範例：

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

Consumer mode 會優先使用 `stytch.user-id`。如果沒有設定，則傳送 `stytch.user-id-prefix + <FNS UID>` 給 Stytch。

## 啟用步驟

1. 透過 HTTPS 部署 FNS，確保公開 base URL 穩定，例如 `https://fns.example.com`。
2. 設定 `server.ext-api-url` 為公開 origin。
3. 設定 Stytch Connected App，並收集 domain、project ID、secret，以及 B2B 情境下的 organization/member IDs。
4. 新增 `oauth` 設定區塊。`resource` 設定為 `https://fns.example.com/api/mcp`，`issuer`、`jwks-url` 與 `authorization-servers` 設定為 Stytch customer domain 及其 JWKS URL。
5. 安全儲存 Stytch secret。使用你的常規 secret 管理方式，或使用未提交的本地設定檔，不要提交真實 Stytch secret。
6. 重啟 FNS 服務。
7. 驗證 protected resource metadata：

   ```bash
   curl -sS https://fns.example.com/.well-known/oauth-protected-resource/api/mcp
   ```

8. 驗證 MCP challenge：

   ```bash
   curl -i https://fns.example.com/api/mcp
   ```

   預期返回 `401`，並包含指向 metadata 的 `WWW-Authenticate` header。

9. 在 ChatGPT 或其他支援 OAuth 的 MCP 用戶端中新增 MCP server：`https://fns.example.com/api/mcp`。
10. 授權過程中，如果 authorize page 要求登入 FNS，請先登入 FNS，因為 bridge endpoints 使用現有 FNS WebGUI token 保護。
11. consent 完成後，確認 MCP 用戶端可以列出並呼叫 tools。

## Static token 相容

如果 `allow-static-fns-token: true`，現有 MCP 用戶端仍可繼續使用 FNS static token：

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

檢查：

- `oauth.enabled` 為 true。
- `https://fns.example.com/.well-known/oauth-protected-resource/api/mcp` 返回 HTTP 200。
- `https://fns.example.com/api/mcp` 返回 `401`，且 `WWW-Authenticate` header 包含 `resource_metadata`。
- `oauth.resource` 與公開 MCP endpoint URL 完全一致。
- ingress 暴露了 `/.well-known/oauth-protected-resource/api/mcp`。

### Metadata 缺少 `authorization_servers`

檢查 `oauth.authorization-servers`。它必須至少包含一個授權伺服器 origin。

### Token verification 返回 `invalid_token`

檢查 token 的 `iss` claim、`oauth.jwks-url` 可達性、JWKS 中是否存在簽名 key、token 是否過期，以及 `oauth.audience` 或 `oauth.resource` 是否符合 token audience/resource。

### Token verification 返回 `insufficient_scope`

檢查 token 是否包含 `oauth.required-scopes` 中列出的 scopes、Stytch Connected App 是否允許簽發請求的 FNS scopes，以及 `default-fns-scope` 是否符合預期。

### WebGUI 沒有 OIDC 登入按鈕

這是預期行為。當前實作不提供 WebGUI OIDC SSO，只保護 MCP routes，並為 MCP 用戶端提供 OAuth authorization bridge。

### `/api/user/auth/callback` 返回 404

這是預期行為。該 endpoint 沒有實作。

### Stytch authorize start 或 submit 失敗

檢查 `oauth.stytch.enabled`、`oauth.stytch.domain`、`project-id`、`secret`、B2B 的 `organization-id` 和 `member-id`、Connected App 的 `client_id`/`redirect_uri`/scopes，以及使用者是否已在 `/oauth/authorize` 前登入 FNS。
