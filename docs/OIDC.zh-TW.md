# WebGUI OIDC 登入

用途：本文說明如何為 Fast Note Sync WebGUI 啟用 OpenID Connect (OIDC) 登入。當你希望使用者透過 Dex、Keycloak、Casdoor 等外部身分提供者登入 WebGUI 時閱讀本文。本文不涵蓋 MCP OAuth 資源伺服器授權；MCP OAuth 使用 `oauth` 配置。

## 功能目標

`oidc` 配置只用於 WebGUI SSO 登入。

啟用後：

- WebGUI 登入頁會請求 `/api/user/auth/oidc/config`；
- 如果服務端啟用了 OIDC，登入頁會顯示配置的 OIDC 登入按鈕；
- `/api/user/auth/oidc/start` 建立 state、nonce 和 PKCE verifier，然後跳轉到身分提供者；
- 身分提供者回呼到 `oidc.redirect-url`；
- 服務端驗證 `id_token`，將 OIDC subject 對應到本機 FNS 使用者，並簽發正常的 WebGUI 登入 token。

## 配置

在 `config/config.yaml` 中加入 `oidc` 配置：

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

當 `enabled: true` 時必須配置：

- `issuer`
- `client-id`
- `client-secret`
- `redirect-url`

預設值：

- `display-name`: `Login with OIDC`
- `callback-path`: `/api/user/auth/oidc/callback`
- `scopes`: `openid`, `profile`, `email`
- `subject-claim`: `sub`
- `email-claim`: `email`
- `username-claim`: `preferred_username`
- `display-name-claim`: `name`

不要把真實的 `client-secret` 提交到公開 Git 配置中。

## 使用者對應

FNS 會把 OIDC 綁定關係儲存在 `user_oidc_identity` 表中。

登入解析順序：

1. 如果 `(issuer, subject)` 已經綁定，直接登入對應本機使用者。
2. 如果沒有綁定，但 OIDC email 符合既有本機使用者，則建立綁定並登入該使用者。
3. 如果沒有符合使用者且 `auto-register: true`，FNS 會建立本機使用者，然後建立綁定。
4. 如果沒有符合使用者且 `auto-register: false`，登入失敗。

較穩妥的上線方式是先設定 `auto-register: false`，預先建立本機使用者，讓首次 OIDC 登入透過 email 自動綁定。

## Provider 配置

### Dex

建立 confidential client：

- Client ID: `fns-webgui`
- Client secret: 與 `oidc.client-secret` 一致
- Redirect URI: `https://fns.example.com/api/user/auth/oidc/callback`
- Scopes: `openid`, `profile`, `email`

`oidc.issuer` 使用 Dex issuer，例如：

```yaml
issuer: "https://dex.example.com/dex"
```

### Keycloak

建立 OpenID Connect confidential client：

- Client ID: `fns-webgui`
- Client authentication: enabled
- Standard flow: enabled
- Valid redirect URI: `https://fns.example.com/api/user/auth/oidc/callback`
- PKCE: 支援 `S256`

`oidc.issuer` 使用 realm issuer：

```yaml
issuer: "https://keycloak.example.com/realms/fns"
```

### Casdoor

建立或更新 application：

- Redirect URI: `https://fns.example.com/api/user/auth/oidc/callback`
- Grant type: `authorization_code`
- Client ID 和 secret 與 `oidc.client-id`、`oidc.client-secret` 一致
- Scopes: `openid`, `profile`, `email`

`oidc.issuer` 使用 Casdoor 對外地址：

```yaml
issuer: "https://casdoor.example.com"
```

Casdoor 常見顯示名稱 claim 是 `displayName`，如需對應可配置：

```yaml
user-mapping:
  display-name-claim: "displayName"
```

## 公網地址與反向代理

`redirect-url` 必須是身分提供者和瀏覽器都能存取到的外部 callback URL。部署在反向代理後面時，應使用公開 HTTPS 地址，而不是容器內部地址。

範例：

```yaml
redirect-url: "https://notes.example.com/api/user/auth/oidc/callback"
```

如果 WebGUI 使用獨立連接埠，callback 仍屬於 API 路由。Provider 中應配置能存取 FNS service 的 callback URL。

## 驗證

倉庫提供 Docker smoke test：

```bash
scripts/oidc-smoke-test.sh
```

它會在本機啟動 Dex、Keycloak、Casdoor，並驗證 provider 相容性。

常規測試不會啟動 Docker：

```bash
go test ./...
```

provider smoke test 內部使用 build tag：

```bash
go test -tags oidc_integration ./internal/oidc -run TestOIDCIntegrationProvider
```

## 排錯

- `oidc provider discovery failed`：檢查 `oidc.issuer` 以及 `/.well-known/openid-configuration`。
- `OIDC state is invalid or expired`：重新開始登入；callback 被重複使用、已過期，或來自另一個服務實例。
- `OIDC token exchange failed`：檢查 client ID、client secret、redirect URL 和 PKCE 支援。
- Provider 登入成功但 FNS 登入失敗：檢查 `email`、`sub` claims，以及是否需要開啟 `auto-register`。
- 登入頁沒有 OIDC 按鈕：檢查 `oidc.enabled: true`，並確認 WebGUI 能以 `X-Client: WebGui` 請求 `/api/user/auth/oidc/config`。
