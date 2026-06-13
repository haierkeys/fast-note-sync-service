# WebGUI OIDC 登入

用途：本文說明如何為 Fast Note Sync WebGUI 啟用 OpenID Connect (OIDC) 登入。當你希望使用者透過 Dex、Keycloak、Casdoor 等外部身分提供者登入 WebGUI 時閱讀本文。本文不涵蓋 MCP OAuth 資源伺服器授權；MCP OAuth 使用 `oauth` 配置。

## 功能目標

`oidc` 配置只用於 WebGUI SSO 登入。

啟用後：

- WebGUI 登入頁會請求 `/api/user/auth/oidc/config`；
- 如果服務端啟用了 OIDC，登入頁會為單一 provider 顯示一個 OIDC 登入按鈕，或為多個 provider 分別顯示按鈕；
- `/api/user/auth/oidc/start` 建立 state、nonce 和 PKCE verifier，然後跳轉到選定的身分提供者；
- 身分提供者回呼到對應配置的 `redirect-url`；
- 服務端驗證 `id_token`，將 OIDC subject 對應到本機 FNS 使用者，並簽發正常的 WebGUI 登入 token。

## 配置

新的 WebGUI 部署建議在 `oidc.providers` 下配置一個或多個 provider：

```yaml
oidc:
  enabled: true
  callback-path: "/api/user/auth/oidc/callback"
  auto-register: false
  user-mapping:
    subject-claim: "sub"
    email-claim: "email"
    username-claim: "preferred_username"
    display-name-claim: "name"
  providers:
    - id: "dex"
      display-name: "Login with Dex"
      issuer: "https://dex.example.com/dex"
      client-id: "fns-webgui"
      client-secret: "change-me"
      redirect-url: "https://fns.example.com/api/user/auth/oidc/callback"
      scopes:
        - openid
        - profile
        - email
    - id: "casdoor"
      display-name: "Login with Casdoor"
      issuer: "https://casdoor.example.com"
      client-id: "fns-webgui"
      client-secret: "change-me"
      redirect-url: "https://fns.example.com/api/user/auth/oidc/callback"
      scopes:
        - openid
        - profile
        - email
      user-mapping:
        display-name-claim: "displayName"
    - id: "keycloak"
      display-name: "Login with Keycloak"
      issuer: "https://keycloak.example.com/realms/fns"
      client-id: "fns-webgui"
      client-secret: "change-me"
      redirect-url: "https://fns.example.com/api/user/auth/oidc/callback"
```

每個 provider 都有一個穩定的 `id`。建議只使用小寫字母、數字和連字號，並保持穩定，讓 provider 選擇和 callback 處理保持可預測。

每個 provider 必須配置：

- `issuer`
- `client-id`
- `client-secret`
- `redirect-url`

Provider 層級的 `user-mapping` 只覆蓋與全域 `oidc.user-mapping` 不同的 claim。例如大多數 provider 使用 `name`，但 Casdoor 需要使用 `displayName` 時，可以只在 Casdoor provider 下覆蓋顯示名稱 claim。

預設值：

- `display-name`: `Login with OIDC`
- `callback-path`: `/api/user/auth/oidc/callback`
- `scopes`: `openid`, `profile`, `email`
- `subject-claim`: `sub`
- `email-claim`: `email`
- `username-claim`: `preferred_username`
- `display-name-claim`: `name`

不要把真實的 `client-secret` 提交到公開 Git 配置中。

### 向後相容的單 Provider 配置

既有的單 provider 部署可以繼續使用歷史的頂層 provider 欄位：

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

這種形式等同於 `oidc.providers` 中只有一個 provider。需要多個 WebGUI 登入按鈕時，優先使用 `providers`。

## 使用者對應

FNS 會把 OIDC 綁定關係儲存在 `user_oidc_identity` 表中。

登入解析順序：

1. 如果 `(issuer, subject)` 已經綁定，直接登入對應本機使用者。
2. 如果沒有綁定，但 OIDC email 符合既有本機使用者，則建立綁定並登入該使用者。
3. 如果沒有符合使用者且 `auto-register: true`，FNS 會建立本機使用者，然後建立綁定。
4. 如果沒有符合使用者且 `auto-register: false`，登入失敗。

較穩妥的上線方式是先設定 `auto-register: false`，預先建立本機使用者，讓首次 OIDC 登入透過 email 自動綁定。

當 `auto-register: true` 時，本機使用者名稱會按以下順序從第一個可用值生成：

1. `username-claim`，例如 `preferred_username`
2. `display-name-claim`，例如 `name`
3. email 中 `@` 前面的部分
4. `oidc_` 加 OIDC subject

生成值會正規化為 FNS 使用者名稱格式：字母、數字、底線，長度 3 到 20。如果使用者名稱已存在，FNS 會追加數字後綴。

## Provider 配置

WebGUI OIDC 登入使用標準 OIDC discovery、authorization code flow、PKCE 和 `id_token` 驗證。Google、Microsoft Entra ID、Auth0、Okta、Zitadel 等 provider 只要提供標準 OIDC issuer、client ID、client secret、redirect URL，並回傳與使用者對應相容的 claims，通常都可以使用。

GitHub 不同：GitHub OAuth Apps 是 OAuth 2.0 provider，並不以同樣方式提供普通 OIDC 登入所需的 discovery 和 `id_token`。如果要使用 GitHub 登入，通常應透過 Dex、Keycloak 或 Casdoor 做 OIDC broker，或使用單獨的 OAuth adapter 把 GitHub OAuth 轉換成需要的登入流程。

### Dex

建立 confidential client：

- Client ID: `fns-webgui`
- Client secret: 與 provider 的 `client-secret` 一致
- Redirect URI: `https://fns.example.com/api/user/auth/oidc/callback`
- Scopes: `openid`, `profile`, `email`

Provider `issuer` 使用 Dex issuer，例如：

```yaml
providers:
  - id: "dex"
    display-name: "Login with Dex"
    issuer: "https://dex.example.com/dex"
    client-id: "fns-webgui"
    client-secret: "change-me"
    redirect-url: "https://fns.example.com/api/user/auth/oidc/callback"
```

### Keycloak

建立 OpenID Connect confidential client：

- Client ID: `fns-webgui`
- Client authentication: enabled
- Standard flow: enabled
- Valid redirect URI: `https://fns.example.com/api/user/auth/oidc/callback`
- PKCE: 支援 `S256`

Provider `issuer` 使用 realm issuer：

```yaml
providers:
  - id: "keycloak"
    display-name: "Login with Keycloak"
    issuer: "https://keycloak.example.com/realms/fns"
    client-id: "fns-webgui"
    client-secret: "change-me"
    redirect-url: "https://fns.example.com/api/user/auth/oidc/callback"
```

### Casdoor

建立或更新 application：

- Redirect URI: `https://fns.example.com/api/user/auth/oidc/callback`
- Grant type: `authorization_code`
- Client ID 和 secret 與 provider 的 `client-id`、`client-secret` 一致
- Scopes: `openid`, `profile`, `email`

Provider `issuer` 使用 Casdoor 對外地址：

```yaml
providers:
  - id: "casdoor"
    display-name: "Login with Casdoor"
    issuer: "https://casdoor.example.com"
    client-id: "fns-webgui"
    client-secret: "change-me"
    redirect-url: "https://fns.example.com/api/user/auth/oidc/callback"
```

Casdoor 常見顯示名稱 claim 是 `displayName`，如需對應可配置：

```yaml
providers:
  - id: "casdoor"
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

- `oidc provider discovery failed`：檢查 provider 的 `issuer` 以及 `/.well-known/openid-configuration`。
- `OIDC state is invalid or expired`：重新開始登入；callback 被重複使用、已過期，或來自另一個服務實例。
- `OIDC token exchange failed`：檢查 client ID、client secret、redirect URL 和 PKCE 支援。
- Provider 登入成功但 FNS 登入失敗：檢查 `email`、`sub` claims，以及是否需要開啟 `auto-register`。
- 登入頁沒有 OIDC 按鈕：檢查 `oidc.enabled: true`，並確認 WebGUI 能以 `X-Client: WebGui` 請求 `/api/user/auth/oidc/config`。
